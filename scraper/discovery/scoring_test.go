package discovery

import (
	"fmt"
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

// dummyKeys builds n distinct URL key strings for use in scoring function tests
// that don't need real URL values (each link gets a unique path so no dedup occurs).
func dummyKeys(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("https://example.com/r/%d", i+1)
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

func TestScoreURLConsistency_FlatNavVsFlatRecipeSlugs(t *testing.T) {
	navHrefs := mustParseURLs(t, []string{
		"https://example.com/about",
		"https://example.com/jobs",
		"https://example.com/press",
		"https://example.com/terms",
	})
	assert.Equal(t, 0.0, urlConsistencyScore(navHrefs), "flat unhyphenated nav chrome should score 0")

	recipeHrefs := mustParseURLs(t, []string{
		"https://example.com/borts-z-purykom",
		"https://example.com/pan-kotlet",
		"https://example.com/sup-iz-hrybamy",
	})
	assert.Equal(t, 0.25, urlConsistencyScore(recipeHrefs), "flat hyphenated recipe slugs should get structural score 0.25")
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

	score := imageDensityScore(links, dummyKeys(len(links)))
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

	score := textQualityScore(links, dummyKeys(len(links)))
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
	deep := mustParseURLs(t, []string{
		"https://example.com/recipes/main/pasta/weeknight",
		"https://example.com/recipes/side/salad/summer",
	})
	deepScore := urlConsistencyScore(deep)
	assert.InDelta(t, 0.5, deepScore, 0.01, "deep path with same prefix depth should score equal to shallow")
}

func TestImageDensityScore_SiblingImage(t *testing.T) {
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

	score := imageDensityScore(links, dummyKeys(len(links)))
	assert.InDelta(t, 1.0, score, 0.01, "sibling images in <li> should be detected")
}

func TestImageDensityScore_NonWrapperParentIgnored(t *testing.T) {
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

	score := imageDensityScore(links, dummyKeys(len(links)))
	assert.InDelta(t, 0.0, score, 0.01, "page-level images should not leak into link density score")
}

func TestCollectGroups_ExternalLinksExcluded(t *testing.T) {
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

func TestMergeSiblingGroups_CollapsesRecipeTiles(t *testing.T) {
	const html = `<html><body><main><section>
		<div><a href="/recipes/pasta">Creamy Pasta Bake</a></div>
		<div><a href="/recipes/chicken">Roast Chicken Dinner</a></div>
		<div><a href="/recipes/salad">Summer Salad Bowl</a></div>
		<div><a href="/recipes/soup">Tomato Basil Soup</a></div>
	</section></main></body></html>`

	doc := parseDocument(t, html)
	base, _ := url.Parse("https://example.com/recipes")
	groups := collectGroups(doc, base)

	require.Greater(t, len(groups), 1, "sibling divs should form separate groups before merge")

	mergeSiblingGroups(groups, ScoringOptions{})

	// Find the group with the most links — it should hold all 4.
	maxLinks := 0
	for _, g := range groups {
		if len(g.links) > maxLinks {
			maxLinks = len(g.links)
		}
	}
	assert.Equal(t, 4, maxLinks, "merged group should contain all sibling links")
}

func TestMergeSiblingGroups_RespectsMinSiblingsToMergeOverride(t *testing.T) {
	const html = `<html><body><main><section>
		<div><a href="/recipes/pasta">Creamy Pasta Bake</a></div>
		<div><a href="/recipes/chicken">Roast Chicken Dinner</a></div>
		<div><a href="/recipes/salad">Summer Salad Bowl</a></div>
		<div><a href="/recipes/soup">Tomato Basil Soup</a></div>
	</section></main></body></html>`

	// 1. With MinSiblingsToMerge set to 5 (greater than the 4 sibling groups), no merge should occur.
	{
		doc := parseDocument(t, html)
		base, _ := url.Parse("https://example.com/recipes")
		groups := collectGroups(doc, base)
		countBefore := len(groups)
		require.Equal(t, 4, countBefore)

		mergeSiblingGroups(groups, ScoringOptions{MinSiblingsToMerge: 5})
		assert.Equal(t, countBefore, len(groups), "should not merge when candidate sibling groups are fewer than MinSiblingsToMerge")
	}

	// 2. With MinSiblingsToMerge set to 3 (at or below the 4 sibling groups), merging should occur.
	{
		doc := parseDocument(t, html)
		base, _ := url.Parse("https://example.com/recipes")
		groups := collectGroups(doc, base)

		mergeSiblingGroups(groups, ScoringOptions{MinSiblingsToMerge: 3})
		maxLinks := 0
		for _, g := range groups {
			if len(g.links) > maxLinks {
				maxLinks = len(g.links)
			}
		}
		assert.Equal(t, 4, maxLinks, "should merge when candidate sibling groups are at least MinSiblingsToMerge")
	}
}

func TestMergeSiblingGroups_PreservesDistinctContainers(t *testing.T) {
	const html = `<html><body><main>
		<ul>
			<li><a href="/recipes/pasta">Pasta</a></li>
			<li><a href="/recipes/chicken">Chicken</a></li>
			<li><a href="/recipes/salad">Salad</a></li>
		</ul>
		<aside>
			<a href="/other/a">Other A</a>
			<a href="/other/b">Other B</a>
			<a href="/other/c">Other C</a>
		</aside>
	</main></body></html>`

	doc := parseDocument(t, html)
	base, _ := url.Parse("https://example.com/recipes")
	groups := collectGroups(doc, base)
	countBefore := len(groups)
	require.GreaterOrEqual(t, countBefore, 2, "should start with at least two distinct groups")

	mergeSiblingGroups(groups, ScoringOptions{})

	assert.Equal(t, countBefore, len(groups), "unrelated containers should not be merged together")
}

// TestCollectGroups_DistinguishesUnrelatedContainersAtSameTruncatedDepth
// covers a recipe list and an unrelated legal-links list in sibling regions
// ("section.hero" vs "aside.promo"), each wrapped in exactly two generically-
// indexed <div>s, so both containers' nearest keyDepth=4 ancestors share the
// same tag/nth-of-type shape and must still resolve to distinct groups.
func TestCollectGroups_DistinguishesUnrelatedContainersAtSameTruncatedDepth(t *testing.T) {
	const html = `<html><body>
		<section class="hero"><div><div><section><ul>
			<li><a href="/recipes/r0">Recipe R0 With A Longer Title</a></li>
			<li><a href="/recipes/r1">Recipe R1 With A Longer Title</a></li>
			<li><a href="/recipes/r2">Recipe R2 With A Longer Title</a></li>
			<li><a href="/recipes/r3">Recipe R3 With A Longer Title</a></li>
		</ul></section></div></div></section>
		<aside class="promo"><div><div><section><ul>
			<li><a href="/legal/doc0">Doc0</a></li>
			<li><a href="/legal/doc1">Doc1</a></li>
			<li><a href="/legal/doc2">Doc2</a></li>
			<li><a href="/legal/doc3">Doc3</a></li>
		</ul></section></div></div></aside>
	</body></html>`

	doc := parseDocument(t, html)
	base, _ := url.Parse("https://example.com/")
	groups := collectGroups(doc, base)

	require.Len(t, groups, 2, "recipe list and legal-links list must not collide into one group")

	for _, g := range groups {
		require.Len(t, g.links, 4)
		firstPath := g.urls[0].Path
		isRecipeGroup := strings.HasPrefix(firstPath, "/recipes/")
		for _, u := range g.urls {
			if isRecipeGroup {
				assert.True(t, strings.HasPrefix(u.Path, "/recipes/"), "recipe group must not contain %q", u.Path)
			} else {
				assert.True(t, strings.HasPrefix(u.Path, "/legal/"), "legal group must not contain %q", u.Path)
			}
		}
	}
}

func TestPickSampleURLs_EvenSpread(t *testing.T) {
	urls := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	got := pickSampleURLs(urls, 3)
	assert.Len(t, got, 3)
	assert.Equal(t, "a", got[0], "first sample should be first URL")
	assert.Equal(t, "j", got[2], "last sample should be last URL")
}

func TestPickSampleURLs_EdgeCases(t *testing.T) {
	assert.Nil(t, pickSampleURLs(nil, 3))
	assert.Nil(t, pickSampleURLs([]string{"a", "b"}, 0))
	assert.Equal(t, []string{"a"}, pickSampleURLs([]string{"a"}, 5))
	got := pickSampleURLs([]string{"a", "b", "c"}, 10)
	assert.Equal(t, []string{"a", "b", "c"}, got, "requesting more than len returns all URLs")
}

func TestPickSampleURLs_NoUnderSampling(t *testing.T) {
	for length := 2; length <= 100; length++ {
		urls := make([]string, length)
		for i := range urls {
			urls[i] = fmt.Sprintf("url-%d", i)
		}
		for n := 1; n < length; n++ {
			got := pickSampleURLs(urls, n)
			assert.Len(t, got, n, "should return exactly n elements for length %d and n %d", length, n)

			// Verify uniqueness.
			seen := make(map[string]bool)
			for _, u := range got {
				if seen[u] {
					t.Fatalf("duplicate URL %s found in sample for length %d and n %d", u, length, n)
				}
				seen[u] = true
			}
		}
	}
}

func TestApplyGroupValidation_PreservesStubImageWhenScrapedHasNone(t *testing.T) {
	stubURL := "https://example.com/recipes/pasta"
	byUrl := map[string]*model.Recipe{
		stubURL: {
			Url:    stubURL,
			Name:   "Creamy Pasta",
			Images: []*model.ImageObject{{Url: "https://example.com/img/pasta.jpg"}},
		},
	}

	// Validator confirms the URL but the scraped page returned no images.
	validate := func(_ []string) []*model.Recipe {
		return []*model.Recipe{{Url: stubURL, Name: "Creamy Pasta Bake"}}
	}

	validCount, sampleCount := applyGroupValidation(byUrl, []string{stubURL}, validate, 1)
	assert.Equal(t, 1, validCount)
	assert.Equal(t, 1, sampleCount)

	entry := byUrl[stubURL]
	assert.Equal(t, "Creamy Pasta Bake", entry.Name, "scraped name should replace stub name")
	require.NotEmpty(t, entry.Images, "stub image must be preserved when scraped recipe has none")
	assert.Equal(t, "https://example.com/img/pasta.jpg", entry.Images[0].Url)
}

func TestApplyGroupValidation_PreservesStubNameWhenScrapedHasNone(t *testing.T) {
	stubURL := "https://example.com/recipes/pasta"
	byUrl := map[string]*model.Recipe{
		stubURL: {Url: stubURL, Name: "Pasta From Link Text"},
	}

	// Validator confirms the URL but the scraped page returned no name.
	validate := func(_ []string) []*model.Recipe {
		return []*model.Recipe{{Url: stubURL}}
	}

	applyGroupValidation(byUrl, []string{stubURL}, validate, 1)

	assert.Equal(t, "Pasta From Link Text", byUrl[stubURL].Name,
		"stub name must be preserved when scraped recipe has no name")
}

func TestApplyGroupValidation_OnlyMergesConfirmedEntries(t *testing.T) {
	urls := []string{
		"https://example.com/a",
		"https://example.com/b",
		"https://example.com/c",
	}
	byUrl := map[string]*model.Recipe{
		urls[0]: {Url: urls[0], Name: "Alpha"},
		urls[1]: {Url: urls[1], Name: "Beta"},
		urls[2]: {Url: urls[2], Name: "Gamma"},
	}

	// Validator confirms only the middle URL.
	validate := func(_ []string) []*model.Recipe {
		return []*model.Recipe{{Url: urls[1], Name: "Beta Confirmed"}}
	}

	validCount, sampleCount := applyGroupValidation(byUrl, urls, validate, 3)
	assert.Equal(t, 1, validCount)
	assert.Equal(t, 3, sampleCount)

	assert.Equal(t, "Alpha", byUrl[urls[0]].Name, "non-confirmed stub must be unchanged")
	assert.Equal(t, "Beta Confirmed", byUrl[urls[1]].Name, "confirmed entry must have merged data")
	assert.Equal(t, "Gamma", byUrl[urls[2]].Name, "non-confirmed stub must be unchanged")
}

func TestReplayDOMScoring_SelectorGone_ReturnsError(t *testing.T) {
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
