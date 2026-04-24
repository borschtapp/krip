package discovery

import (
	"fmt"
	"log"
	"math"
	"net/url"
	"slices"
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

// exclusionSelector matches non-content zones to ignore.
const exclusionSelector = "nav, header, footer, [role=navigation], [role=banner], [role=contentinfo], .sidebar, #sidebar, .nav, #nav, .header, #header, .footer, #footer"

// imgSrcAttrs lists attributes that may hold an image URL, including lazy-load variants.
var imgSrcAttrs = []string{"src", "data-src", "data-lazy-src", "data-lazy", "srcset"}

// imgSrc returns the first non-empty image attribute from the selection.
func imgSrc(sel *goquery.Selection) string {
	for _, attr := range imgSrcAttrs {
		if v, ok := sel.Attr(attr); ok && v != "" {
			return v
		}
	}
	return ""
}

// linkImg returns the <img> nearest to a link, checking inside <a> and its immediate parent.
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
		return fmt.Errorf("%s requires a parsed document", SourceDOMContainer)
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

// maxGroupsToValidate caps the number of candidate groups to check.
const maxGroupsToValidate = 3

// tryDOMScoring finds and validates the best-scoring container of recipe links.
func tryDOMScoring(data *model.DataInput, feed *model.Feed, baseUrl *url.URL, sampling SamplingOptions) (*model.DiscoveredFeed, error) {
	if data.Document == nil {
		return nil, fmt.Errorf("discovery: no document")
	}

	groups := collectGroups(data.Document, baseUrl)
	mergeSiblingGroups(groups)
	for k, g := range groups {
		if len(g.links) < minGroupSize {
			delete(groups, k)
		}
	}
	if len(groups) == 0 {
		return nil, fmt.Errorf("discovery: no candidate containers found")
	}

	ranked := rankGroups(groups)
	if ranked[0].score < sampleThreshold {
		return nil, fmt.Errorf("discovery: best container score %.2f below threshold", ranked[0].score)
	}

	sg, byUrl, urls, err := selectBestGroup(ranked, baseUrl, sampling)
	if err != nil {
		return nil, err
	}

	for _, u := range urls {
		feed.AddEntry(byUrl[u])
	}
	return makeDiscoveredFeed(sg, urls), nil
}

// selectBestGroup iterates ranked candidates, optionally validates each, and returns
// the best group together with its deduplicated entry map and URL slice.
func selectBestGroup(ranked []scoredGroup, baseUrl *url.URL, sampling SamplingOptions) (scoredGroup, map[string]*model.Recipe, []string, error) {
	type fallbackCandidate struct {
		byUrl      map[string]*model.Recipe
		urls       []string
		sg         scoredGroup
		validCount int
	}
	var best *fallbackCandidate

	for i, sg := range ranked {
		if sg.score < sampleThreshold || (sampling.Validator != nil && i >= maxGroupsToValidate) {
			break
		}

		byUrl, urls := buildGroupEntries(sg.group, baseUrl)
		validCount, sampleCount := applyGroupValidation(byUrl, urls, sampling.Validator, sampling.SampleSize)

		if sampleCount == 0 || validCount*2 > sampleCount {
			return sg, byUrl, urls, nil
		}

		log.Printf("discovery: group %q failed validation (%d/%d recipe URLs), trying next", sg.group.key, validCount, sampleCount)

		if best == nil || validCount > best.validCount {
			best = &fallbackCandidate{byUrl, urls, sg, validCount}
		}
	}

	if best != nil && best.validCount > 0 {
		log.Printf("discovery: using best-effort group %q with %d confirmed recipe URL(s)", best.sg.group.key, best.validCount)
		return best.sg, best.byUrl, best.urls, nil
	}

	return scoredGroup{}, nil, nil, fmt.Errorf("discovery: no candidate group passed validation")
}

func makeDiscoveredFeed(sg scoredGroup, urls []string) *model.DiscoveredFeed {
	return &model.DiscoveredFeed{
		Source:          SourceDOMContainer,
		Selector:        sg.group.key,
		UrlPattern:      UrlPathPattern(urls),
		ConfidenceScore: sg.score,
	}
}

// buildGroupEntries creates a deduplicated map of recipes from a group's links.
func buildGroupEntries(group *containerGroup, baseUrl *url.URL) (byUrl map[string]*model.Recipe, urls []string) {
	byUrl = make(map[string]*model.Recipe, len(group.links))
	for _, a := range group.links {
		href, _ := a.Attr("href")
		abs := utils.ToAbsoluteUrl(baseUrl, href)
		entry, seen := byUrl[abs]
		if !seen {
			urls = append(urls, abs)
			entry = &model.Recipe{Url: abs}
			byUrl[abs] = entry
		}
		if entry.Name == "" {
			if text := linkVisibleText(a); text != "" {
				entry.Name = utils.CleanupInline(text)
			}
		}
		if len(entry.Images) == 0 {
			if src := imgSrc(linkImg(a)); src != "" {
				entry.AddImageUrl(utils.ToAbsoluteUrl(baseUrl, src))
			}
		}
	}
	return byUrl, urls
}

// applyGroupValidation samples URLs and merges confirmed recipe data back.
func applyGroupValidation(byUrl map[string]*model.Recipe, urls []string, validate GroupValidator, sampleSize int) (validCount, sampleCount int) {
	if validate == nil || sampleSize == 0 {
		return 0, 0
	}

	sample := pickSampleURLs(urls, sampleSize)
	validated := validate(sample)

	for _, v := range validated {
		if e, ok := byUrl[v.Url]; ok {
			// Preserve stub fields as fallback if scraping misses them.
			savedName := e.Name
			savedImages := e.Images
			*e = *v
			if e.Name == "" {
				e.Name = savedName
			}
			if len(e.Images) == 0 {
				e.Images = savedImages
			}
		}
	}
	return len(validated), len(sample)
}

// extractFromPattern returns all links whose paths start with the pattern.
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

// isExcluded returns true if the selection is within an ignored zone.
func isExcluded(s *goquery.Selection) bool {
	return s.ParentsFiltered(exclusionSelector).Length() > 0
}

// singleLinkWrappers are tags that typically wrap a single <a>.
var singleLinkWrappers = map[string]bool{
	"li": true, "dt": true, "dd": true, "td": true, "th": true,
}

// effectiveParent walks up from a link's direct parent to find the meaningful
// container, skipping single-link wrapper tags like <li>.
//
// HTML5 semantic grouping: <article> is treated as a card element. When
// encountered, the walk continues up through intermediate <div> wrappers (up to
// keyDepth levels) to reach the nearest semantic container — typically a <section>,
// <main>, <ul>, or <ol>. This ensures that all <article> cards inside the same
// <section> are assigned to the section's group rather than forming per-article
// micro-groups that require sibling merging.
func effectiveParent(a *goquery.Selection) (*goquery.Selection, string) {
	cur := a.Parent()
	passedArticle := false
	articleDivDepth := 0
	for cur.Length() > 0 {
		n := cur.Get(0)
		if n == nil || n.Type != html.ElementNode {
			cur = cur.Parent()
			continue
		}
		if singleLinkWrappers[n.Data] {
			cur = cur.Parent()
			continue
		}
		if n.Data == "article" {
			passedArticle = true
			articleDivDepth = 0
			cur = cur.Parent()
			continue
		}
		// After passing an <article>, skip intermediate <div> wrappers to reach the
		// nearest semantic container (section, main, ul, ol, …). Stop at the first
		// non-div element — that is the natural grouping parent for the card layout.
		if passedArticle && n.Data == "div" && articleDivDepth < keyDepth {
			articleDivDepth++
			cur = cur.Parent()
			continue
		}
		return cur, n.Data
	}
	p := a.Parent()
	if n := p.Get(0); n != nil {
		return p, n.Data
	}
	return p, ""
}

// collectGroups scans the document for links and groups them by containerKey.
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
		// Skip external links and the page itself.
		if absURL.Host != baseUrl.Host ||
			(absURL.Path == baseUrl.Path && absURL.RawQuery == baseUrl.RawQuery) {
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

	return groups
}

// viableCandidate is a merge candidate with enough sibling groups.
type viableCandidate struct {
	prefix     string
	childTag   string
	uniqueKeys []string
}

// mergeSiblingGroups collapses groups whose container keys share the same ancestor
// prefix when the varying sibling index is stripped out. This handles card/grid
// layouts where each recipe tile is a sibling <div> wrapped in additional single-
// child divs, so links inside tiles end up in separate groups instead of one
// shared container.
//
// To handle multi-level wrapping (e.g. grid > tile:nth(N) > inner:nth(1) > <a>),
// the function tries all possible strip depths (1 to keyDepth-1). It always
// prefers the DEEPEST (most specific) common ancestor that has at least
// minGroupSize sibling child groups, so over-merging is minimised.
func mergeSiblingGroups(groups map[string]*containerGroup) {
	// candidateAncestor represents a proposed merge: a common ancestor prefix
	// and the set of child group keys that would be merged under it.
	type candidateAncestor struct {
		prefix   string
		childTag string
		keys     []string
	}

	// prefixToCandidate maps each possible ancestor prefix to its candidate.
	// We accumulate child keys from all groups for each prefix.
	prefixToCandidate := map[string]*candidateAncestor{}

	for k := range groups {
		parts := strings.Split(k, " > ")
		// Try stripping 1, 2, …, len(parts)-1 trailing segments.
		for stripDepth := 1; stripDepth < len(parts); stripDepth++ {
			prefix := strings.Join(parts[:len(parts)-stripDepth], " > ")
			// The tag of the merged group = the tag of the stripped segment right
			// above the common prefix (i.e. the first segment still in the prefix).
			childSegment := parts[len(parts)-stripDepth]
			childTag, _, _ := strings.Cut(childSegment, ":")
			if c, ok := prefixToCandidate[prefix]; ok {
				c.keys = append(c.keys, k)
				if semanticBonus(childTag) > semanticBonus(c.childTag) {
					c.childTag = childTag
				}
			} else {
				prefixToCandidate[prefix] = &candidateAncestor{prefix: prefix, childTag: childTag, keys: []string{k}}
			}
		}
	}

	// Collect all viable candidates (enough siblings), then sort deepest-first
	// (most specific ancestor = smallest over-merge risk).
	var viable []viableCandidate
	for _, c := range prefixToCandidate {
		// Deduplicate child keys (a group may have been registered at multiple depths).
		uniqueKeys := utils.Deduplicate(c.keys)
		if len(uniqueKeys) >= minGroupSize {
			viable = append(viable, viableCandidate{c.prefix, c.childTag, uniqueKeys})
		}
	}

	// Sort by depth descending (deepest first).
	slices.SortFunc(viable, func(a, b viableCandidate) int {
		da := strings.Count(a.prefix, " > ")
		db := strings.Count(b.prefix, " > ")
		if da != db {
			return db - da
		}
		return len(b.prefix) - len(a.prefix)
	})

	// Greedily merge from deepest to shallowest.
	consumed := map[string]bool{}
	for _, c := range viable {
		var pendingKeys []string
		for _, k := range c.uniqueKeys {
			if !consumed[k] {
				pendingKeys = append(pendingKeys, k)
			}
		}
		if len(pendingKeys) < minGroupSize {
			continue
		}
		merged := &containerGroup{key: c.prefix, tag: c.childTag}
		for _, ck := range pendingKeys {
			g := groups[ck]
			merged.links = append(merged.links, g.links...)
			merged.urls = append(merged.urls, g.urls...)
			consumed[ck] = true
		}
		if existing, ok := groups[c.prefix]; ok {
			existing.links = append(existing.links, merged.links...)
			existing.urls = append(existing.urls, merged.urls...)
		} else {
			groups[c.prefix] = merged
		}
	}
	for k := range consumed {
		delete(groups, k)
	}
}

// scoredGroup pairs a container group with its computed heuristic score.
type scoredGroup struct {
	group *containerGroup
	score float64
}

// rankGroups returns all groups sorted by descending heuristic score.
func rankGroups(groups map[string]*containerGroup) []scoredGroup {
	ranked := make([]scoredGroup, 0, len(groups))
	for _, g := range groups {
		ranked = append(ranked, scoredGroup{g, score(g)})
	}
	slices.SortFunc(ranked, func(a, b scoredGroup) int {
		if a.score > b.score {
			return -1
		}
		if a.score < b.score {
			return 1
		}
		return 0
	})
	return ranked
}

// pickSampleURLs returns up to n URLs spread evenly across the slice.
func pickSampleURLs(urls []string, n int) []string {
	if n <= 0 || len(urls) == 0 {
		return nil
	}
	if n == 1 {
		return urls[:1]
	}
	if n >= len(urls) {
		return append([]string(nil), urls...)
	}
	result := make([]string, n)
	for i := range result {
		result[i] = urls[i*(len(urls)-1)/(n-1)]
	}
	return utils.Deduplicate(result)
}

// deduplicateGroupURLs returns per-link string keys and a deduplicated URL slice
// in a single pass. keys holds one string per link (from url.String()); uniq holds
// original *url.URL pointers without duplicates.
// Multi-link cards (e.g. image <a> + title <a> pointing to the same URL) collapse
// to one entry in uniq while keys retains a key for every link index.
func deduplicateGroupURLs(g *containerGroup) (keys []string, uniq []*url.URL) {
	keys = make([]string, len(g.urls))
	seen := make(map[string]struct{}, len(g.urls))
	for i, u := range g.urls {
		k := u.String()
		keys[i] = k
		if _, ok := seen[k]; !ok {
			seen[k] = struct{}{}
			uniq = append(uniq, u)
		}
	}
	return
}

// score computes the weighted composite score for a container group.
//
// URL consistency weight (0.30) is intentionally lower than text quality (0.25)
// because many recipe sites use root-level slugs that share no common prefix,
// while content-taxonomy sections (ingredients, categories) may have a prefix but
// contain non-recipe content. Long descriptive titles are a stronger recipe signal.
func score(g *containerGroup) float64 {
	keys, uniq := deduplicateGroupURLs(g)
	return 0.30*urlConsistencyScore(uniq) +
		0.25*imageDensityScore(g.links, keys) +
		0.25*textQualityScore(g.links, keys) +
		0.10*countScore(len(uniq)) +
		0.10*semanticBonus(g.tag)
}

// urlConsistencyScore measures how uniform the hrefs are (common path prefix).
// Uses prefixLen/(prefixLen+1) so score is depth-independent: 1-seg→0.5, 2-seg→0.67, 3-seg→0.75.
// When there is no common prefix but all paths share the same depth (e.g. all root-level
// slugs: /recipe-a/, /recipe-b/), a partial structural-consistency score of 0.25 is returned.
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
		// No shared prefix: check if all paths have the same depth (structural consistency).
		// Sites like klopotenko.com use root-level slugs (/recipe-slug/) that share no prefix
		// but are structurally uniform — this is still a positive signal.
		depth := len(allSegments[0])
		if depth > 0 {
			for _, segs := range allSegments[1:] {
				if len(segs) != depth {
					return 0
				}
			}
			return 0.25
		}
		return 0
	}
	return float64(prefixLen) / float64(prefixLen+1)
}

// imageDensityScore measures what fraction of unique URLs have a nearby <img>.
// Deduplicates by URL so that sites with two <a> tags per card (one image-only,
// one title-only) are scored correctly: a URL counts as having an image if ANY
// of its links contains one.
// Checks inside the <a> first; falls back to the immediate parent when it is
// a single-link wrapper tag (e.g. <li>), handling sibling-image patterns like:
//
//	<li><img src="photo.jpg"><a href="/recipes/x">Title</a></li>
//
// keys must be pre-computed string representations of each link's URL (parallel to links).
func imageDensityScore(links []*goquery.Selection, keys []string) float64 {
	if len(links) == 0 || len(keys) == 0 {
		return 0
	}
	urlHasImage := make(map[string]bool, len(keys))
	for i, a := range links {
		key := keys[i]
		if !urlHasImage[key] {
			urlHasImage[key] = imgSrc(linkImg(a)) != ""
		}
	}
	hits := 0
	for _, v := range urlHasImage {
		if v {
			hits++
		}
	}
	return float64(hits) / float64(len(urlHasImage))
}

// linkVisibleText returns link text, excluding SSR-injected <style> or <script>.
func linkVisibleText(a *goquery.Selection) string {
	if a.Find("style, script").Length() == 0 {
		return strings.TrimSpace(a.Text())
	}
	clone := a.Clone()
	clone.Find("style, script").Remove()
	return strings.TrimSpace(clone.Text())
}

// textSegmentScore returns the quality score for a single word count.
func textSegmentScore(words int) float64 {
	switch {
	case words >= 3 && words <= 8:
		return 1.0
	case (words >= 1 && words <= 2) || (words >= 9 && words <= 12):
		return 0.5
	default:
		return 0
	}
}

// textQualityScore measures whether link text looks like a recipe title (3–8 words).
// Deduplicates by URL so that sites with two <a> tags per card (one image-only,
// one title-only) are scored correctly: each unique URL takes the best text score
// across all of its links.
// Modern SPA cards often concatenate title + metadata (separated by "|" or "•") into
// a single text string. We therefore score each "|"/"•"-delimited segment and take
// the best, so a card with "Great Pasta | 30 min | Easy" scores as "Great Pasta".
//
// keys must be pre-computed string representations of each link's URL (parallel to links).
func textQualityScore(links []*goquery.Selection, keys []string) float64 {
	if len(links) == 0 || len(keys) == 0 {
		return 0
	}
	urlBestScore := make(map[string]float64, len(keys))
	for i, a := range links {
		key := keys[i]
		text := linkVisibleText(a)
		s := textSegmentScore(len(strings.Fields(text)))
		if s < 1.0 {
			for _, seg := range strings.FieldsFunc(text, func(r rune) bool { return r == '|' || r == '•' || r == '·' }) {
				if ss := textSegmentScore(len(strings.Fields(strings.TrimSpace(seg)))); ss > s {
					s = ss
				}
			}
		}
		if cur, seen := urlBestScore[key]; !seen || s > cur {
			urlBestScore[key] = s
		}
	}
	total := 0.0
	for _, s := range urlBestScore {
		total += s
	}
	return total / float64(len(urlBestScore))
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

// UrlPathPattern returns the common path prefix of URLs as a string.
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
	return strings.FieldsFunc(path, func(r rune) bool { return r == '/' })
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
