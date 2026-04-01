package discovery

import (
	"fmt"
	"math"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/utils"
)

const (
	keyDepth        = 4    // ancestor levels used to build containerKey
	minGroupSize    = 3    // groups smaller than this are ignored
	sampleThreshold = 0.55 // accept only with sampling
	// AcceptThreshold is the confidence score above which results are accepted without sampling.
	AcceptThreshold = 0.70
)

// exclusionSelector matches navigation/chrome zones whose links should be ignored.
const exclusionSelector = "nav, header, footer, [role=navigation], [role=banner], [role=contentinfo], .sidebar, #sidebar, .nav, #nav, .header, #header, .footer, #footer"

// imgSrcAttrs lists the HTML attributes that may hold an image URL, in priority order.
// Many recipe sites lazy-load images via data-* attributes.
var imgSrcAttrs = []string{"src", "data-src", "data-lazy-src", "data-lazy", "srcset"}

// imgSrc returns the first non-empty image src-like attribute value from sel,
// checking standard and common lazy-load attributes.
func imgSrc(sel *goquery.Selection) string {
	for _, attr := range imgSrcAttrs {
		if v, ok := sel.Attr(attr); ok && v != "" {
			return v
		}
	}
	return ""
}

// linkImg returns the <img> nearest to a link: checks inside <a> first, then
// falls back to the immediate parent when it is a single-link wrapper (e.g. <li>)
// to handle sibling-image patterns like <li><img><a href="..."></li>.
func linkImg(a *goquery.Selection) *goquery.Selection {
	img := a.Find("img").First()
	if img.Length() == 0 {
		parent := a.Parent()
		if n := parent.Get(0); n != nil && singleLinkWrappers[n.Data] {
			img = parent.Find("img").First()
		}
	}
	return img
}

// containerGroup holds the links belonging to one container.
type containerGroup struct {
	key   string
	links []*goquery.Selection
	urls  []*url.URL // resolved absolute URLs, parallel to links
	tag   string     // direct parent tag
}

func replayDOMScoring(data *model.DataInput, feed *model.Feed, d *model.DiscoveredFeed) error {
	if data.Document == nil {
		return fmt.Errorf("dom-container requires a parsed document")
	}
	baseUrl, err := url.Parse(data.Url)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}

	if d.Selector == "" {
		for _, u := range extractFromPattern(data, baseUrl, d.UrlPattern) {
			feed.AddEntry(&model.Recipe{Url: u})
		}
		return nil
	}

	container := data.Document.Find(d.Selector)
	if container.Length() == 0 {
		return fmt.Errorf("dom-container: selector %q no longer exists in document", d.Selector)
	}

	seen := map[string]struct{}{}
	container.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		abs := utils.ToAbsoluteUrl(baseUrl, href)
		if d.UrlPattern != "" && !utils.UrlMatchesPathPattern(abs, d.UrlPattern) {
			return
		}
		if _, dup := seen[abs]; dup {
			return
		}
		seen[abs] = struct{}{}
		feed.AddEntry(&model.Recipe{Url: abs})
	})

	if len(feed.Entries) == 0 {
		return fmt.Errorf("dom-container: selector %q yielded no entries", d.Selector)
	}
	return nil
}

// tryDOMScoring finds the highest-scoring container of recipe links,
// adds stub Recipe entries to feed, and returns the DiscoveredFeed descriptor.
func tryDOMScoring(data *model.DataInput, feed *model.Feed, baseUrl *url.URL) (*model.DiscoveredFeed, error) {
	if data.Document == nil {
		return nil, fmt.Errorf("discovery: no document")
	}

	groups := collectGroups(data.Document, baseUrl)

	if len(groups) == 0 {
		return nil, fmt.Errorf("discovery: no candidate containers found")
	}

	best, score := pickBestGroup(groups)
	if best == nil || score < sampleThreshold {
		return nil, fmt.Errorf("discovery: best container score %.2f below threshold", score)
	}

	seen := map[string]struct{}{}
	var urls []string
	for _, a := range best.links {
		href, _ := a.Attr("href")
		abs := utils.ToAbsoluteUrl(baseUrl, href)
		if _, dup := seen[abs]; dup {
			continue
		}
		seen[abs] = struct{}{}
		urls = append(urls, abs)

		entry := &model.Recipe{Url: abs}
		if text := strings.TrimSpace(a.Text()); text != "" {
			entry.Name = utils.CleanupInline(text)
		}
		if src := imgSrc(linkImg(a)); src != "" {
			entry.AddImageUrl(utils.ToAbsoluteUrl(baseUrl, src))
		}
		feed.AddEntry(entry)
	}

	return &model.DiscoveredFeed{
		Source:          SourceDOMContainer,
		Selector:        best.key,
		UrlPattern:      UrlPathPattern(urls),
		ConfidenceScore: score,
	}, nil
}

// extractFromPattern scans all <a href> elements in data.Document and returns
// those whose path starts with the given prefix. Used for fast replay.
func extractFromPattern(data *model.DataInput, baseUrl *url.URL, pattern string) []string {
	if data.Document == nil || pattern == "" {
		return nil
	}

	var urls []string
	data.Document.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		abs := utils.ToAbsoluteUrl(baseUrl, href)
		if utils.UrlMatchesPathPattern(abs, pattern) {
			urls = append(urls, abs)
		}
	})
	return urls
}

// isExcluded returns true if sel is a descendant of a navigation/chrome zone.
func isExcluded(s *goquery.Selection) bool {
	return s.ParentsFiltered(exclusionSelector).Length() > 0
}

// singleLinkWrappers are tags that typically wrap a single <a> and should be
// skipped when looking for the meaningful group container.
var singleLinkWrappers = map[string]bool{
	"li": true, "dt": true, "dd": true, "td": true, "th": true,
}

// effectiveParent walks up from a link's direct parent to find the meaningful
// container, skipping single-link wrapper tags like <li>.
func effectiveParent(a *goquery.Selection) (*goquery.Selection, string) {
	cur := a.Parent()
	for cur.Length() > 0 {
		n := cur.Get(0)
		if n == nil || n.Type != html.ElementNode {
			cur = cur.Parent()
			continue
		}
		if !singleLinkWrappers[n.Data] {
			return cur, n.Data
		}
		cur = cur.Parent()
	}
	p := a.Parent()
	if n := p.Get(0); n != nil {
		return p, n.Data
	}
	return p, ""
}

// collectGroups gathers all internal <a href> elements (outside exclusion zones),
// groups them by their effective container's structural key.
func collectGroups(doc *goquery.Document, baseUrl *url.URL) map[string]*containerGroup {
	groups := make(map[string]*containerGroup)
	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if href == "" || strings.HasPrefix(href, "#") {
			return
		}
		u, err := url.Parse(href)
		if err != nil {
			return
		}
		absURL := baseUrl.ResolveReference(u)
		// Skip external links and the page itself (compare hosts, not string prefix,
		// to avoid matching https://example.com.evil.com/ as same-origin).
		if absURL.Host != baseUrl.Host || absURL.String() == baseUrl.String() {
			return
		}
		if isExcluded(s) {
			return
		}

		parent, parentTag := effectiveParent(s)
		if parent.Length() == 0 {
			return
		}
		key := containerKey(parent, keyDepth)

		if g, ok := groups[key]; ok {
			g.links = append(g.links, s)
			g.urls = append(g.urls, absURL)
		} else {
			groups[key] = &containerGroup{key: key, links: []*goquery.Selection{s}, urls: []*url.URL{absURL}, tag: parentTag}
		}
	})

	// Discard groups that are too small
	for k, g := range groups {
		if len(g.links) < minGroupSize {
			delete(groups, k)
		}
	}
	return groups
}

// pickBestGroup returns the group with the highest composite score.
func pickBestGroup(groups map[string]*containerGroup) (*containerGroup, float64) {
	var best *containerGroup
	bestScore := -1.0

	for _, g := range groups {
		s := score(g)
		if s > bestScore {
			bestScore = s
			best = g
		}
	}
	return best, bestScore
}

// score computes the weighted composite score for a container group.
func score(g *containerGroup) float64 {
	return 0.35*urlConsistencyScore(g.urls) +
		0.25*imageDensityScore(g.links) +
		0.20*textQualityScore(g.links) +
		0.10*countScore(len(g.links)) +
		0.10*semanticBonus(g.tag)
}

// urlConsistencyScore measures how uniform the hrefs are (common path prefix).
// Uses prefixLen/(prefixLen+1) so score is depth-independent: 1-seg→0.5, 2-seg→0.67, 3-seg→0.75.
func urlConsistencyScore(urls []*url.URL) float64 {
	if len(urls) == 0 {
		return 0
	}
	var allSegments [][]string
	for _, u := range urls {
		segs := pathSegments(u.Path)
		allSegments = append(allSegments, segs)
	}
	if len(allSegments) == 0 {
		return 0
	}
	prefixLen := longestCommonPrefixLength(allSegments)
	if prefixLen == 0 {
		return 0
	}
	return float64(prefixLen) / float64(prefixLen+1)
}

// imageDensityScore measures what fraction of links have a nearby <img>.
// Checks inside the <a> first; falls back to the immediate parent when it is
// a single-link wrapper tag (e.g. <li>), handling sibling-image patterns like:
//
//	<li><img src="photo.jpg"><a href="/recipes/x">Title</a></li>
func imageDensityScore(links []*goquery.Selection) float64 {
	if len(links) == 0 {
		return 0
	}
	hits := 0
	for _, a := range links {
		if imgSrc(linkImg(a)) != "" {
			hits++
		}
	}
	return float64(hits) / float64(len(links))
}

// textQualityScore measures whether link text looks like a recipe title (3–8 words).
func textQualityScore(links []*goquery.Selection) float64 {
	if len(links) == 0 {
		return 0
	}
	total := 0.0
	for _, a := range links {
		words := len(strings.Fields(strings.TrimSpace(a.Text())))
		switch {
		case words >= 3 && words <= 8:
			total += 1.0
		case (words >= 1 && words <= 2) || (words >= 9 && words <= 12):
			total += 0.5
		}
	}
	return total / float64(len(links))
}

// countScore rewards larger groups (20+ links = 1.0).
func countScore(n int) float64 {
	return math.Min(float64(n)/20.0, 1.0)
}

// semanticBonus rewards links inside semantic HTML5 containers.
func semanticBonus(tag string) float64 {
	switch tag {
	case "main", "article", "section", "ul", "ol":
		return 1.0
	case "div":
		return 0.5
	default:
		return 0.0
	}
}

// containerKey builds a structural path string by walking up to depth ancestors.
// Example: "main:nth-of-type(1) > ul:nth-of-type(1)"
func containerKey(sel *goquery.Selection, depth int) string {
	if depth == 0 || sel.Length() == 0 {
		return ""
	}
	n := sel.Get(0)
	if n == nil || n.Type != html.ElementNode {
		return ""
	}

	tag := n.Data
	idx := 1
	for sib := n.PrevSibling; sib != nil; sib = sib.PrevSibling {
		if sib.Type == html.ElementNode && sib.Data == tag {
			idx++
		}
	}
	part := fmt.Sprintf("%s:nth-of-type(%d)", tag, idx)

	parent := sel.Parent()
	if parent.Length() == 0 {
		return part
	}
	parentKey := containerKey(parent, depth-1)
	if parentKey == "" {
		return part
	}
	return parentKey + " > " + part
}

// UrlPathPattern returns the common path prefix of the given URLs as a plain string.
// Example: ["/recipes/pasta", "/recipes/chicken"] → "/recipes/"
func UrlPathPattern(urls []string) string {
	if len(urls) == 0 {
		return ""
	}
	var allSegs [][]string
	for _, u := range urls {
		parsed, err := url.Parse(u)
		if err != nil {
			continue
		}
		allSegs = append(allSegs, pathSegments(parsed.Path))
	}
	if len(allSegs) == 0 {
		return ""
	}
	prefixLen := longestCommonPrefixLength(allSegs)
	if prefixLen == 0 {
		return ""
	}
	return "/" + strings.Join(allSegs[0][:prefixLen], "/") + "/"
}

// pathSegments splits a URL path into non-empty segments.
func pathSegments(path string) []string {
	var segs []string
	for _, s := range strings.Split(path, "/") {
		if s != "" {
			segs = append(segs, s)
		}
	}
	return segs
}

// longestCommonPrefixLength returns the number of leading segments shared by all slices.
func longestCommonPrefixLength(segs [][]string) int {
	if len(segs) == 0 {
		return 0
	}
	minLen := len(segs[0])
	for _, s := range segs[1:] {
		if len(s) < minLen {
			minLen = len(s)
		}
	}
	for i := 0; i < minLen; i++ {
		seg := segs[0][i]
		for _, s := range segs[1:] {
			if s[i] != seg {
				return i
			}
		}
	}
	return minLen
}
