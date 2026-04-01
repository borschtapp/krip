package discovery

import (
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/borschtapp/krip/model"
)

func mustParseURLs(t *testing.T, strs []string) []*url.URL {
	t.Helper()
	out := make([]*url.URL, 0, len(strs))
	for _, s := range strs {
		u, err := url.Parse(s)
		require.NoError(t, err)
		out = append(out, u)
	}
	return out
}

func parseDocument(t *testing.T, html string) *goquery.Document {
	t.Helper()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)
	return doc
}

func TestContainerKey_Stability(t *testing.T) {
	const html = `<html><body><main><ul><li><a href="/r/1">Recipe One</a></li></ul></main></body></html>`

	doc := parseDocument(t, html)

	var key1, key2 string
	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		key1 = containerKey(s.Parent().Parent(), keyDepth)
	})
	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		key2 = containerKey(s.Parent().Parent(), keyDepth)
	})

	assert.Equal(t, key1, key2, "containerKey must be deterministic")
	assert.NotEmpty(t, key1)
}

func TestContainerKey_IncludesTag(t *testing.T) {
	const html = `<html><body><main><ul><li><a href="/r/1">Recipe</a></li></ul></main></body></html>`
	doc := parseDocument(t, html)
	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		key := containerKey(s.Parent(), keyDepth)
		assert.Contains(t, key, "li")
	})
}

func TestScoreURLConsistency_HighScore(t *testing.T) {
	hrefs := mustParseURLs(t, []string{
		"https://example.com/recipes/pasta",
		"https://example.com/recipes/chicken",
		"https://example.com/recipes/stew",
	})
	score := urlConsistencyScore(hrefs)
	assert.GreaterOrEqual(t, score, 0.5, "consistent /recipes/ prefix should score high")
}

func TestScoreURLConsistency_LowScore(t *testing.T) {
	hrefs := mustParseURLs(t, []string{
		"https://example.com/recipes/pasta",
		"https://example.com/blog/post-1",
		"https://example.com/about",
	})
	score := urlConsistencyScore(hrefs)
	assert.Less(t, score, 0.5, "mixed paths should score low")
}

func TestImageDensityScore(t *testing.T) {
	const html = `<html><body>
		<a href="/r/1"><img src="/i/1.jpg"><span>Recipe One</span></a>
		<a href="/r/2"><img src="/i/2.jpg"><span>Recipe Two</span></a>
		<a href="/r/3"><span>No Image</span></a>
	</body></html>`

	doc := parseDocument(t, html)
	var links []*goquery.Selection
	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		links = append(links, s)
	})

	score := imageDensityScore(links)
	assert.InDelta(t, 2.0/3.0, score, 0.01)
}

func TestTextQualityScore_RecipeTitles(t *testing.T) {
	const html = `<html><body>
		<a href="/r/1">Classic Pasta Carbonara</a>
		<a href="/r/2">Roast Chicken with Herbs</a>
		<a href="/r/3">X</a>
	</body></html>`

	doc := parseDocument(t, html)
	var links []*goquery.Selection
	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		links = append(links, s)
	})

	score := textQualityScore(links)
	// 2 good titles (3-8 words = 1.0 each), 1 bad (1 word = 0.5) → avg = (1+1+0.5)/3 ≈ 0.83
	assert.Greater(t, score, 0.7)
}

func TestCountScore(t *testing.T) {
	assert.Equal(t, 0.0, countScore(0))
	assert.InDelta(t, 0.5, countScore(10), 0.01)
	assert.Equal(t, 1.0, countScore(20))
	assert.Equal(t, 1.0, countScore(50))
}

func TestSemanticBonus(t *testing.T) {
	assert.Equal(t, 1.0, semanticBonus("main"))
	assert.Equal(t, 1.0, semanticBonus("ul"))
	assert.Equal(t, 1.0, semanticBonus("section"))
	assert.Equal(t, 0.5, semanticBonus("div"))
	assert.Equal(t, 0.0, semanticBonus("span"))
}

func TestLongestCommonPrefixLength(t *testing.T) {
	tests := []struct {
		segs     [][]string
		expected int
	}{
		{[][]string{{"recipes", "pasta"}, {"recipes", "chicken"}}, 1},
		{[][]string{{"a", "b", "c"}, {"a", "b", "d"}}, 2},
		{[][]string{{"a"}, {"b"}}, 0},
		{[][]string{}, 0},
	}
	for _, tc := range tests {
		got := longestCommonPrefixLength(tc.segs)
		assert.Equal(t, tc.expected, got)
	}
}

func TestScoreURLConsistency_DepthIndependent(t *testing.T) {
	// Shallow path: 1-segment prefix on 2-segment URLs → 1/(1+1) = 0.5
	shallow := mustParseURLs(t, []string{
		"https://example.com/recipes/pasta",
		"https://example.com/recipes/chicken",
	})
	shallowScore := urlConsistencyScore(shallow)
	assert.InDelta(t, 0.5, shallowScore, 0.01, "shallow path should score 0.5")

	// Deep path: 1-segment prefix on 4-segment URLs → also 1/(1+1) = 0.5
	// (old formula: 1/4 = 0.25, penalising deep paths unfairly)
	deep := mustParseURLs(t, []string{
		"https://example.com/recipes/main/pasta/weeknight",
		"https://example.com/recipes/side/salad/summer",
	})
	deepScore := urlConsistencyScore(deep)
	assert.InDelta(t, 0.5, deepScore, 0.01, "deep path with same prefix depth should score equal to shallow")
}

func TestImageDensityScore_SiblingImage(t *testing.T) {
	// Images are siblings of <a> inside <li>, not children of <a>
	const html = `<html><body>
		<ul>
			<li><img src="/i/1.jpg"><a href="/r/1">Recipe One</a></li>
			<li><img src="/i/2.jpg"><a href="/r/2">Recipe Two</a></li>
			<li><img src="/i/3.jpg"><a href="/r/3">Recipe Three</a></li>
		</ul>
	</body></html>`

	doc := parseDocument(t, html)
	var links []*goquery.Selection
	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		links = append(links, s)
	})

	score := imageDensityScore(links)
	assert.InDelta(t, 1.0, score, 0.01, "sibling images in <li> should be detected")
}

func TestImageDensityScore_NonWrapperParentIgnored(t *testing.T) {
	// Images are elsewhere on the page, not in <li> wrapper — should NOT be counted
	const html = `<html><body>
		<img src="/header.jpg">
		<a href="/r/1">Recipe One</a>
		<a href="/r/2">Recipe Two</a>
		<a href="/r/3">Recipe Three</a>
	</body></html>`

	doc := parseDocument(t, html)
	var links []*goquery.Selection
	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		links = append(links, s)
	})

	score := imageDensityScore(links)
	assert.InDelta(t, 0.0, score, 0.01, "page-level images should not leak into link density score")
}

func TestCollectGroups_ExternalLinksExcluded(t *testing.T) {
	// SSRF/pollution: a link whose host looks like the page host (prefix match) but isn't
	const html = `<html><body><main><ul>
		<li><a href="/recipes/pasta">Pasta</a></li>
		<li><a href="/recipes/chicken">Chicken</a></li>
		<li><a href="/recipes/salad">Salad</a></li>
		<li><a href="https://example.com.evil.com/recipes/hijack">Evil</a></li>
		<li><a href="https://other.com/recipes/foreign">Foreign</a></li>
	</ul></main></body></html>`

	doc := parseDocument(t, html)
	base, _ := url.Parse("https://example.com/recipes")
	groups := collectGroups(doc, base)

	for _, g := range groups {
		for _, a := range g.links {
			href, _ := a.Attr("href")
			assert.NotContains(t, href, "evil.com", "external link should be excluded from scoring groups")
			assert.NotContains(t, href, "other.com", "external link should be excluded from scoring groups")
		}
	}
}

func TestImgSrc_LazyLoadAttrs(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		wantSrc bool
	}{
		{"standard src", `<img src="/img/pasta.jpg">`, true},
		{"data-src", `<img data-src="/img/pasta.jpg">`, true},
		{"data-lazy-src", `<img data-lazy-src="/img/pasta.jpg">`, true},
		{"data-lazy", `<img data-lazy="/img/pasta.jpg">`, true},
		{"srcset", `<img srcset="/img/pasta.jpg 1x">`, true},
		{"no src attrs", `<img alt="placeholder">`, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc := parseDocument(t, "<html><body>"+tc.html+"</body></html>")
			img := doc.Find("img").First()
			src := imgSrc(img)
			if tc.wantSrc {
				assert.NotEmpty(t, src)
			} else {
				assert.Empty(t, src)
			}
		})
	}
}

func TestUrlPathPattern(t *testing.T) {
	urls := []string{
		"https://example.com/recipes/pasta",
		"https://example.com/recipes/chicken",
		"https://example.com/recipes/stew",
	}
	pattern := UrlPathPattern(urls)
	assert.NotEmpty(t, pattern)
	assert.Contains(t, pattern, "recipes")
}

func TestReplayDOMScoring_UsesSelector(t *testing.T) {
	const htmlPage = `<html><body>
		<nav><a href="/about">About</a></nav>
		<main><ul>
			<li><a href="/recipes/pasta">Pasta</a></li>
			<li><a href="/recipes/chicken">Chicken</a></li>
			<li><a href="/recipes/stew">Stew</a></li>
		</ul></main>
	</body></html>`

	doc := parseDocument(t, htmlPage)
	data := &model.DataInput{Url: "https://example.com", Document: doc}
	feed := &model.Feed{}
	d := &model.DiscoveredFeed{
		Source:     SourceDOMContainer,
		Selector:   "main:nth-of-type(1) > ul:nth-of-type(1)",
		UrlPattern: "/recipes/",
	}

	err := replayDOMScoring(data, feed, d)

	require.NoError(t, err)
	assert.Len(t, feed.Entries, 3)
	// The /about link in <nav> must not appear.
	for _, e := range feed.Entries {
		assert.Contains(t, e.Url, "/recipes/")
	}
}

func TestReplayDOMScoring_SelectorGone_ReturnsError(t *testing.T) {
	// Page no longer has the originally discovered container.
	const htmlPage = `<html><body><div><a href="/recipes/pasta">Pasta</a></div></body></html>`

	doc := parseDocument(t, htmlPage)
	data := &model.DataInput{Url: "https://example.com", Document: doc}
	feed := &model.Feed{}
	d := &model.DiscoveredFeed{
		Source:   SourceDOMContainer,
		Selector: "main:nth-of-type(1) > ul:nth-of-type(1)", // no longer present
	}

	err := replayDOMScoring(data, feed, d)

	assert.Error(t, err, "missing container should return an error to trigger re-discovery")
	assert.Len(t, feed.Entries, 0)
}
