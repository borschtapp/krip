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
	"sync"

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

	// filterRecipeLocs already de-duplicates by loc, and feed is known-empty here
	// (replay is the first entry-finding stage), so appending directly is
	// equivalent to feed.AddEntry but skips its per-call O(len(feed.Entries)) scan
	// — significant when a sitemap yields thousands of locs.
	for _, loc := range filterRecipeLocs(locs) {
		if d.UrlPattern != "" && !utils.UrlMatchesPathPattern(loc, d.UrlPattern) {
			continue
		}
		feed.Entries = append(feed.Entries, &model.Recipe{Url: loc})
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
		// recipeLocs is already unique (filterRecipeLocs dedupes) and feed starts
		// empty, so appending directly avoids AddEntry's O(len(feed.Entries)) scan
		// per loc — this loop is exactly the "tens of thousands of locs" case §13
		// of discovery-ideas.md calls out as quadratic.
		for _, loc := range recipeLocs {
			feed.Entries = append(feed.Entries, &model.Recipe{Url: loc})
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
	seen := make(map[string]struct{})

	addCandidate := func(u string) {
		if u == "" {
			return
		}
		if _, dup := seen[u]; !dup {
			seen[u] = struct{}{}
			candidates = append(candidates, u)
		}
	}

	if data.Document != nil {
		if href, exists := data.Document.Find(`link[rel="sitemap"]`).First().Attr("href"); exists && href != "" {
			addCandidate(utils.ToAbsoluteUrl(baseUrl, href))
		}
	}

	base := utils.BaseUrl(baseUrl.String())
	for _, u := range robotsSitemaps(base, data.RequestOptions) {
		addCandidate(u)
	}

	addCandidate(base + "/sitemap_index.xml")
	addCandidate(base + "/sitemap.xml")

	return candidates
}

// robotsSitemaps fetches robots.txt and returns any "Sitemap:" directive URLs.
// robots.txt is the canonical registry for non-standard sitemap locations
// (/sitemap-index.xml, /wp-sitemap.xml, /sitemap/sitemap.xml, ...) that the
// hardcoded guesses below miss entirely; probing it turns what would otherwise be
// wasted failed probes plus a missed feed into one small extra request.
func robotsSitemaps(base string, opts model.RequestOptions) []string {
	body, _, err := utils.ExecuteRequest(utils.RequestConfig{Method: "GET", URL: base + "/robots.txt"}, opts)
	if err != nil {
		return nil
	}

	const directive = "sitemap:"
	var sitemaps []string
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if idx := strings.IndexByte(line, '#'); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}
		if len(line) <= len(directive) || !strings.EqualFold(line[:len(directive)], directive) {
			continue
		}
		if u := strings.TrimSpace(line[len(directive):]); u != "" {
			sitemaps = append(sitemaps, u)
		}
	}
	return sitemaps
}

// followSitemapIndex fetches child sitemaps from an index, returning all their locs.
// Prefers children whose URL contains recipe-related keywords; fetches at most 3.
func followSitemapIndex(indexLocs []string, opts model.RequestOptions) ([]string, error) {
	// Sort: recipe-keyword children first
	var preferred, rest []string
	for _, loc := range indexLocs {
		if matchesRecipePath(loc) {
			preferred = append(preferred, loc)
		} else {
			rest = append(rest, loc)
		}
	}
	ordered := append(preferred, rest...)

	limit := min(3, len(ordered))

	// Selection (which children to fetch, and in what preference order) already
	// happened above; fetching those <= 3 children concurrently only shortens wall
	// time; it does not change which sitemaps get fetched. Merge order below still
	// follows the original preference order for determinism.
	results := make([][]string, limit)
	var wg sync.WaitGroup
	for i, childUrl := range ordered[:limit] {
		wg.Add(1)
		go func(i int, childUrl string) {
			defer wg.Done()
			body, _, err := utils.ExecuteRequest(utils.RequestConfig{Method: "GET", URL: childUrl}, opts)
			if err != nil {
				return
			}
			body, _ = maybeDecompress(body)
			locs, isIndex, err := parseSitemap(body)
			if err != nil || isIndex {
				return
			}
			results[i] = locs
		}(i, childUrl)
	}
	wg.Wait()

	seen := map[string]struct{}{}
	var allLocs []string
	for _, locs := range results {
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

// recipePathPattern matches recipe-related keywords in a URL path.
// Used both to filter individual recipe page URLs and to prioritise sitemap index children.
var recipePathPattern = regexp.MustCompile(`(?i)(recipes?|cook|food|dish|meal|recettes?|cuisine|rezepte?|kochen|gerichte?|recetas?|cocina|platos?|receitas?|cozinha|pratos?|ricette?|cucina|piatti|recepty?|przepisy?|dania?)`)

// matchesRecipePath matches recipePathPattern against the URL path only, not the
// whole URL. Matching the whole URL would make every loc on a host like
// "recipes.example.com" or "mycookingblog.com" match regardless of its path,
// silently turning the recipe filter into a no-op on exactly the sites this
// keyword list targets.
func matchesRecipePath(rawUrl string) bool {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return false
	}
	return recipePathPattern.MatchString(u.Path)
}

func filterRecipeLocs(locs []string) []string {
	seen := map[string]struct{}{}
	filtered := make([]string, 0, len(locs))
	for _, loc := range locs {
		if _, dup := seen[loc]; dup {
			continue
		}
		seen[loc] = struct{}{}
		if matchesRecipePath(loc) {
			filtered = append(filtered, loc)
		}
	}
	return filtered
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
