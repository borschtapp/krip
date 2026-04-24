package discovery

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/utils"
)

func replaySitemap(data *model.DataInput, feed *model.Feed, d *model.DiscoveredFeed) error {
	if d.Selector == "" {
		return fmt.Errorf("sitemap source has no URL")
	}

	body, _, err := utils.ExecuteRequest(
		utils.RequestConfig{Method: "GET", URL: d.Selector},
		data.RequestOptions,
	)
	if err != nil {
		return fmt.Errorf("sitemap fetch failed: %w", err)
	}

	body, err = maybeDecompress(body)
	if err != nil {
		return fmt.Errorf("sitemap decompress failed: %w", err)
	}
	locs, isIndex, err := parseSitemap(body)
	if err != nil {
		return fmt.Errorf("sitemap parse failed: %w", err)
	}
	if isIndex {
		locs, _ = followSitemapIndex(locs, data.RequestOptions)
	}

	for _, loc := range filterRecipeLocs(locs) {
		if d.UrlPattern != "" && !utils.UrlMatchesPathPattern(loc, d.UrlPattern) {
			continue
		}
		feed.AddEntry(&model.Recipe{Url: loc})
	}

	return nil
}

// trySitemap discovers recipe URLs by fetching and parsing a sitemap.
func trySitemap(data *model.DataInput, feed *model.Feed, baseUrl *url.URL, opts model.RequestOptions) (*model.DiscoveredFeed, error) {
	candidates := sitemapCandidates(data, baseUrl)

	for _, sitemapUrl := range candidates {
		body, _, err := utils.ExecuteRequest(utils.RequestConfig{Method: "GET", URL: sitemapUrl}, opts)
		if err != nil {
			continue
		}

		body, err = maybeDecompress(body)
		if err != nil {
			continue
		}

		locs, isIndex, err := parseSitemap(body)
		if err != nil {
			continue
		}

		if isIndex {
			locs, err = followSitemapIndex(locs, opts)
			if err != nil || len(locs) == 0 {
				continue
			}
		}

		recipeLocs := filterRecipeLocs(locs)
		if len(recipeLocs) < 5 {
			continue
		}

		prefix := UrlPathPattern(recipeLocs)
		for _, loc := range recipeLocs {
			feed.AddEntry(&model.Recipe{Url: loc})
		}
		return &model.DiscoveredFeed{
			Source:     SourceSitemap,
			Selector:   sitemapUrl,
			UrlPattern: prefix,
		}, nil
	}

	return nil, fmt.Errorf("no usable sitemap found")
}

// sitemapCandidates returns sitemap URLs to probe, in priority order.
func sitemapCandidates(data *model.DataInput, baseUrl *url.URL) []string {
	var candidates []string

	if data.Document != nil {
		if href, exists := data.Document.Find(`link[rel="sitemap"]`).First().Attr("href"); exists && href != "" {
			candidates = append(candidates, utils.ToAbsoluteUrl(baseUrl, href))
		}
	}

	base := utils.BaseUrl(baseUrl.String())
	candidates = append(candidates,
		base+"/sitemap_index.xml",
		base+"/sitemap.xml",
	)
	return candidates
}

// followSitemapIndex fetches child sitemaps from an index, returning all their locs.
// Prefers children whose URL contains recipe-related keywords; fetches at most 3.
func followSitemapIndex(indexLocs []string, opts model.RequestOptions) ([]string, error) {
	// Sort: recipe-keyword children first
	var preferred, rest []string
	for _, loc := range indexLocs {
		if containsRecipeKeyword(loc) {
			preferred = append(preferred, loc)
		} else {
			rest = append(rest, loc)
		}
	}
	ordered := append(preferred, rest...)

	limit := min(3, len(ordered))

	seen := map[string]struct{}{}
	var allLocs []string
	for _, childUrl := range ordered[:limit] {
		body, _, err := utils.ExecuteRequest(utils.RequestConfig{Method: "GET", URL: childUrl}, opts)
		if err != nil {
			continue
		}
		body, _ = maybeDecompress(body)
		locs, isIndex, err := parseSitemap(body)
		if err != nil || isIndex {
			continue
		}
		for _, loc := range locs {
			if _, dup := seen[loc]; !dup {
				seen[loc] = struct{}{}
				allLocs = append(allLocs, loc)
			}
		}
	}
	return allLocs, nil
}

// XML structs for stdlib sitemap parsing.
// No namespace is specified in struct tags: Go's xml.Decoder matches by local
// name when the struct tag has no namespace, which handles both
// namespace-qualified sitemaps (xmlns="http://www.sitemaps.org/...") and
// namespace-free sitemaps equally well.
type sitemapXML struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []struct {
		Loc string `xml:"loc"`
	} `xml:"url"`
}

type sitemapIndexXML struct {
	XMLName  xml.Name `xml:"sitemapindex"`
	Sitemaps []struct {
		Loc string `xml:"loc"`
	} `xml:"sitemap"`
}

// parseSitemap parses a sitemap body. Returns (locs, isIndex, error).
// Uses a single-pass decoder: peek at the root element name, then decode
// into the appropriate struct once (avoiding the previous double-unmarshal).
func parseSitemap(body []byte) ([]string, bool, error) {
	dec := xml.NewDecoder(bytes.NewReader(body))
	// Advance to the first start element (skipping processing instructions etc.)
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, false, fmt.Errorf("not a valid sitemap")
		}
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch start.Name.Local {
		case "urlset":
			var urlset sitemapXML
			if err := dec.DecodeElement(&urlset, &start); err != nil {
				return nil, false, fmt.Errorf("not a valid sitemap: %w", err)
			}
			locs := make([]string, 0, len(urlset.URLs))
			for _, u := range urlset.URLs {
				if u.Loc != "" {
					locs = append(locs, u.Loc)
				}
			}
			if len(locs) == 0 {
				return nil, false, fmt.Errorf("not a valid sitemap: urlset is empty")
			}
			return locs, false, nil
		case "sitemapindex":
			var index sitemapIndexXML
			if err := dec.DecodeElement(&index, &start); err != nil {
				return nil, false, fmt.Errorf("not a valid sitemap: %w", err)
			}
			locs := make([]string, 0, len(index.Sitemaps))
			for _, s := range index.Sitemaps {
				if s.Loc != "" {
					locs = append(locs, s.Loc)
				}
			}
			if len(locs) == 0 {
				return nil, false, fmt.Errorf("not a valid sitemap: index is empty")
			}
			return locs, true, nil
		default:
			return nil, false, fmt.Errorf("not a valid sitemap: unexpected root element %q", start.Name.Local)
		}
	}
}

// recipeKeywords is the canonical list of recipe-related keywords used for
// both URL path pattern matching and sitemap index prioritisation.
var recipeKeywords = []string{"recipe", "cook", "food", "dish", "meal"}

// recipePathPatterns matches URL paths that likely point to recipe pages.
var recipePathPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)/recipes?/`),
	regexp.MustCompile(`(?i)/cook/`),
	regexp.MustCompile(`(?i)/food/`),
	regexp.MustCompile(`(?i)/dish/`),
	regexp.MustCompile(`(?i)/meal/`),
}

func filterRecipeLocs(locs []string) []string {
	seen := map[string]struct{}{}
	filtered := make([]string, 0, len(locs))
	for _, loc := range locs {
		if _, dup := seen[loc]; dup {
			continue
		}
		seen[loc] = struct{}{}
		for _, pat := range recipePathPatterns {
			if pat.MatchString(loc) {
				filtered = append(filtered, loc)
				break
			}
		}
	}
	return filtered
}

func containsRecipeKeyword(s string) bool {
	lower := strings.ToLower(s)
	for _, kw := range recipeKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// maxDecompressBytes is the upper bound for decompressed sitemap content.
// Sitemaps over 50 MB uncompressed are almost certainly not recipe indexes.
const maxDecompressBytes = 50 << 20 // 50 MB

// maybeDecompress decompresses gzip data if detected.
// Limits output to maxDecompressBytes to guard against gzip bombs.
func maybeDecompress(data []byte) ([]byte, error) {
	if len(data) < 2 || data[0] != 0x1f || data[1] != 0x8b {
		return data, nil
	}
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = r.Close()
	}()
	return io.ReadAll(io.LimitReader(r, maxDecompressBytes))
}
