package discovery_test

import (
	"os"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/scraper/discovery"
)

func parseDocumentFile(t *testing.T, path string) *goquery.Document {
	t.Helper()
	f, err := os.Open(path)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()
	doc, err := goquery.NewDocumentFromReader(f)
	require.NoError(t, err)
	return doc
}

func parseDocument(t *testing.T, htmlContent string) *goquery.Document {
	t.Helper()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	require.NoError(t, err)
	return doc
}

// TestScrapeFeed_DOMContainer verifies that the DOM container scorer finds the
// main recipe listing and ignores nav/footer links.
func TestScrapeFeed_DOMContainer(t *testing.T) {
	const html = `<!DOCTYPE html><html><body>
		<nav><a href="/about">About</a><a href="/contact">Contact</a></nav>
		<main><ul>
			<li><a href="/recipes/pasta"><img src="/img/pasta.jpg"><span>Creamy Pasta Bake</span></a></li>
			<li><a href="/recipes/chicken"><img src="/img/chicken.jpg"><span>Roast Chicken Dinner</span></a></li>
			<li><a href="/recipes/salad"><img src="/img/salad.jpg"><span>Summer Salad Bowl</span></a></li>
			<li><a href="/recipes/soup"><img src="/img/soup.jpg"><span>Tomato Basil Soup</span></a></li>
			<li><a href="/recipes/steak"><img src="/img/steak.jpg"><span>Pan-Seared Steak</span></a></li>
		</ul></main>
		<footer><a href="/privacy">Privacy</a><a href="/terms">Terms</a></footer>
	</body></html>`

	doc := parseDocument(t, html)
	data := &model.DataInput{Url: "https://example.com/recipes", Document: doc}
	feed := &model.Feed{}

	err := discovery.ScrapeFeed(data, feed, discovery.SamplingOptions{}, discovery.ScoringOptions{})
	require.NoError(t, err)

	assert.NotEmpty(t, feed.Entries)
	assert.NotNil(t, feed.Discovered)
	assert.Equal(t, discovery.SourceDOMContainer, feed.Discovered.Source)
	assert.NotEmpty(t, feed.Discovered.Selector)
	assert.NotEmpty(t, feed.Discovered.UrlPattern)
	assert.Greater(t, feed.Discovered.ConfidenceScore, 0.0)

	// Nav and footer links must not appear
	for _, entry := range feed.Entries {
		assert.Contains(t, entry.Url, "/recipes/")
	}
}

// TestScrapeFeed_ReplayDiscovered verifies that ReplayDiscovered finds all page links
// matching the stored UrlPattern without re-running the full discovery scoring.
func TestScrapeFeed_ReplayDiscovered(t *testing.T) {
	discovered := &model.DiscoveredFeed{
		Source:     discovery.SourceDOMContainer,
		UrlPattern: `/recipes/`,
	}

	const html = `<!DOCTYPE html><html><body><main><ul>
		<li><a href="/recipes/pasta"><span>Pasta</span></a></li>
		<li><a href="/recipes/chicken"><span>Chicken</span></a></li>
		<li><a href="/recipes/new-recipe"><span>Brand New Recipe</span></a></li>
	</ul></main></body></html>`

	doc := parseDocument(t, html)
	data := &model.DataInput{Url: "https://example.com/recipes", Document: doc}
	feed := &model.Feed{}

	err := discovery.ReplayDiscovered(data, feed, discovered)
	require.NoError(t, err)

	urls := make([]string, 0, len(feed.Entries))
	for _, e := range feed.Entries {
		urls = append(urls, e.Url)
	}

	assert.Contains(t, urls, "https://example.com/recipes/pasta")
	assert.Contains(t, urls, "https://example.com/recipes/chicken")
	assert.Contains(t, urls, "https://example.com/recipes/new-recipe")
}

// TestScrapeFeed_ExcludesNavLinks verifies that nav/header/footer links are excluded.
func TestScrapeFeed_ExcludesNavLinks(t *testing.T) {
	const html = `<!DOCTYPE html><html><body>
		<nav>
			<a href="/recipes/nav-item-1">Nav Recipe One</a>
			<a href="/recipes/nav-item-2">Nav Recipe Two</a>
			<a href="/recipes/nav-item-3">Nav Recipe Three</a>
		</nav>
		<main><ul>
			<li><a href="/recipes/pasta"><img src="/img/pasta.jpg"><span>Pasta Carbonara</span></a></li>
			<li><a href="/recipes/chicken"><img src="/img/chicken.jpg"><span>Chicken Tikka</span></a></li>
			<li><a href="/recipes/stew"><img src="/img/stew.jpg"><span>Beef Stew</span></a></li>
			<li><a href="/recipes/tart"><img src="/img/tart.jpg"><span>Lemon Tart</span></a></li>
		</ul></main>
	</body></html>`

	doc := parseDocument(t, html)
	data := &model.DataInput{Url: "https://example.com/recipes", Document: doc}
	feed := &model.Feed{}

	err := discovery.ScrapeFeed(data, feed, discovery.SamplingOptions{}, discovery.ScoringOptions{})
	require.NoError(t, err)

	for _, entry := range feed.Entries {
		assert.NotContains(t, entry.Url, "nav-item",
			"nav link should be excluded from feed entries")
	}
}

// TestScrapeFeed_ReplayStateless verifies that dom-container replay derives URLs from the
// current page only. A URL no longer on the page must not appear in results.
func TestScrapeFeed_ReplayStateless(t *testing.T) {
	discovered := &model.DiscoveredFeed{
		Source:     discovery.SourceDOMContainer,
		UrlPattern: "/recipes/",
	}

	// "pasta" has been removed from the listing; only "brand-new" is present now.
	const html = `<!DOCTYPE html><html><body><main><ul>
		<li><a href="/recipes/brand-new"><span>Brand New Recipe</span></a></li>
	</ul></main></body></html>`

	doc := parseDocument(t, html)
	data := &model.DataInput{Url: "https://example.com/recipes", Document: doc}
	feed := &model.Feed{}

	err := discovery.ReplayDiscovered(data, feed, discovered)
	require.NoError(t, err)

	urls := make([]string, 0, len(feed.Entries))
	for _, e := range feed.Entries {
		urls = append(urls, e.Url)
	}

	assert.Contains(t, urls, "https://example.com/recipes/brand-new")
	assert.NotContains(t, urls, "https://example.com/recipes/pasta",
		"URL removed from page should not appear in stateless replay")
}

// TestScrapeFeed_MetadataExtracted verifies that Name and Image are populated
// from the DOM container during initial discovery.
func TestScrapeFeed_MetadataExtracted(t *testing.T) {
	const html = `<!DOCTYPE html><html><body><main><ul>
		<li><a href="/recipes/pasta"><img src="/img/pasta.jpg"><span>Creamy Pasta Bake</span></a></li>
		<li><a href="/recipes/chicken"><img src="/img/chicken.jpg"><span>Roast Chicken</span></a></li>
		<li><a href="/recipes/salad"><img src="/img/salad.jpg"><span>Summer Salad Bowl</span></a></li>
		<li><a href="/recipes/soup"><img src="/img/soup.jpg"><span>Tomato Basil Soup</span></a></li>
		<li><a href="/recipes/steak"><img src="/img/steak.jpg"><span>Pan-Seared Steak</span></a></li>
	</ul></main></body></html>`

	doc := parseDocument(t, html)
	data := &model.DataInput{Url: "https://example.com/recipes", Document: doc}
	feed := &model.Feed{}

	err := discovery.ScrapeFeed(data, feed, discovery.SamplingOptions{}, discovery.ScoringOptions{})
	require.NoError(t, err)
	require.NotEmpty(t, feed.Entries)

	for _, entry := range feed.Entries {
		assert.NotEmpty(t, entry.Name, "entry %s should have a name extracted from DOM", entry.Url)
		assert.NotEmpty(t, entry.Images, "entry %s should have an image extracted from DOM", entry.Url)
	}
}

// TestScrapeFeed_SiblingImageExtracted verifies that images that are siblings of
// the <a> (not children) are still captured during discovery.
func TestScrapeFeed_SiblingImageExtracted(t *testing.T) {
	// Images are siblings of <a>, inside <li>
	const html = `<!DOCTYPE html><html><body><main><ul>
		<li><img src="/img/pasta.jpg"><a href="/recipes/pasta">Creamy Pasta Bake</a></li>
		<li><img src="/img/chicken.jpg"><a href="/recipes/chicken">Roast Chicken</a></li>
		<li><img src="/img/salad.jpg"><a href="/recipes/salad">Summer Salad</a></li>
		<li><img src="/img/soup.jpg"><a href="/recipes/soup">Tomato Soup</a></li>
		<li><img src="/img/steak.jpg"><a href="/recipes/steak">Pan Steak</a></li>
	</ul></main></body></html>`

	doc := parseDocument(t, html)
	data := &model.DataInput{Url: "https://example.com/recipes", Document: doc}
	feed := &model.Feed{}

	err := discovery.ScrapeFeed(data, feed, discovery.SamplingOptions{}, discovery.ScoringOptions{})
	require.NoError(t, err)
	require.NotEmpty(t, feed.Entries)

	for _, entry := range feed.Entries {
		assert.NotEmpty(t, entry.Images, "entry %s should capture sibling image from <li>", entry.Url)
	}
}

// TestScrapeFeed_EmptyDocument verifies that ScrapeFeed returns an error gracefully.
func TestScrapeFeed_EmptyDocument(t *testing.T) {
	data := &model.DataInput{Url: "https://example.com/recipes", Document: nil}
	feed := &model.Feed{}

	err := discovery.ScrapeFeed(data, feed, discovery.SamplingOptions{}, discovery.ScoringOptions{})
	assert.Error(t, err)
	assert.Empty(t, feed.Entries)
}

// TestScrapeFeed_ScoreThreshold verifies boundary behaviour: a container that
// barely clears sampleThreshold (0.55) is accepted, while one that doesn't is rejected.
func TestScrapeFeed_ScoreThreshold(t *testing.T) {
	// Build a listing that will score just above threshold:
	// consistent /recipes/ URLs, some images, good text
	makeRecipeList := func(count int) string {
		titles := []string{"Classic Pasta Bake", "Roast Chicken Dinner", "Summer Salad Bowl",
			"Tomato Basil Soup", "Pan-Seared Steak", "Lemon Drizzle Cake",
			"Beef Stew", "Prawn Linguine", "Mushroom Risotto", "Apple Crumble"}
		slugs := []string{"pasta", "chicken", "salad", "soup", "steak",
			"cake", "stew", "linguine", "risotto", "crumble"}
		items := ""
		for i := 0; i < count && i < len(titles); i++ {
			items += `<li><a href="/recipes/` + slugs[i] + `"><img src="/img/` + slugs[i] + `.jpg"><span>` + titles[i] + `</span></a></li>`
		}
		return `<!DOCTYPE html><html><body><main><ul>` + items + `</ul></main></body></html>`
	}

	t.Run("accepted above threshold", func(t *testing.T) {
		doc := parseDocument(t, makeRecipeList(8))
		data := &model.DataInput{Url: "https://example.com/recipes", Document: doc}
		feed := &model.Feed{}
		err := discovery.ScrapeFeed(data, feed, discovery.SamplingOptions{}, discovery.ScoringOptions{})
		require.NoError(t, err)
		assert.NotEmpty(t, feed.Entries)
	})

	t.Run("rejected below threshold", func(t *testing.T) {
		// Mixed random links — no consistent pattern, no images, single words
		const html = `<!DOCTYPE html><html><body><main><ul>
			<li><a href="/about">About</a></li>
			<li><a href="/contact">Contact</a></li>
			<li><a href="/blog/post1">Post</a></li>
		</ul></main></body></html>`
		doc := parseDocument(t, html)
		data := &model.DataInput{Url: "https://example.com", Document: doc}
		feed := &model.Feed{}
		err := discovery.ScrapeFeed(data, feed, discovery.SamplingOptions{}, discovery.ScoringOptions{})
		assert.Error(t, err, "low-quality container should be rejected")
	})
}

// TestScrapeFeed_ImageDensityTiebreaker verifies that when two containers have
// equal link counts, the one with images wins.
func TestScrapeFeed_ImageDensityTiebreaker(t *testing.T) {
	const html = `<!DOCTYPE html><html><body><main>
		<section id="with-images">
			<ul>
				<li><a href="/recipes/pasta"><img src="/img/pasta.jpg"><span>Pasta Bake</span></a></li>
				<li><a href="/recipes/chicken"><img src="/img/chicken.jpg"><span>Roast Chicken</span></a></li>
				<li><a href="/recipes/salad"><img src="/img/salad.jpg"><span>Summer Salad</span></a></li>
				<li><a href="/recipes/soup"><img src="/img/soup.jpg"><span>Tomato Soup</span></a></li>
				<li><a href="/recipes/steak"><img src="/img/steak.jpg"><span>Pan Steak</span></a></li>
			</ul>
		</section>
		<section id="without-images">
			<ul>
				<li><a href="/recipes/omelette">Classic Omelette</a></li>
				<li><a href="/recipes/pancakes">Fluffy Pancakes</a></li>
				<li><a href="/recipes/waffles">Belgian Waffles</a></li>
				<li><a href="/recipes/muffins">Blueberry Muffins</a></li>
				<li><a href="/recipes/granola">Homemade Granola</a></li>
			</ul>
		</section>
	</main></body></html>`

	doc := parseDocument(t, html)
	data := &model.DataInput{Url: "https://example.com/recipes", Document: doc}
	feed := &model.Feed{}

	err := discovery.ScrapeFeed(data, feed, discovery.SamplingOptions{}, discovery.ScoringOptions{})
	require.NoError(t, err)

	// All winning entries should come from the with-images section
	// (pasta/chicken/salad/soup/steak), not the without-images section
	entryUrls := make(map[string]bool)
	for _, e := range feed.Entries {
		entryUrls[e.Url] = true
	}
	assert.True(t, entryUrls["https://example.com/recipes/pasta"] ||
		entryUrls["https://example.com/recipes/chicken"],
		"entries from image-rich container should be present")
}

// TestScrapeFeed_ArticleCardsPattern verifies that article-card layouts — where each
// recipe is a top-level <article> child of a <section> — are correctly discovered.
// This exercises the ONE-level sibling merge (vs. the two-level merge in grid_tiles.html),
// data-src lazy-load image detection, and bullet (•) separated metadata scoring.
func TestScrapeFeed_ArticleCardsPattern(t *testing.T) {
	doc := parseDocumentFile(t, "testdata/article_cards.html")
	data := &model.DataInput{
		Url:      "https://foodblog.example.com/recipes",
		Document: doc,
	}
	feed := &model.Feed{}

	err := discovery.ScrapeFeed(data, feed, discovery.SamplingOptions{}, discovery.ScoringOptions{})
	require.NoError(t, err, "ScrapeFeed should succeed on the article-card page")

	assert.Equal(t, discovery.SourceDOMContainer, feed.Discovered.Source)
	assert.GreaterOrEqual(t, feed.Discovered.ConfidenceScore, 0.55)

	// All 12 recipe articles must be discovered.
	require.GreaterOrEqual(t, len(feed.Entries), 12, "all recipe articles should be discovered")

	entryURLs := make(map[string]bool, len(feed.Entries))
	for _, e := range feed.Entries {
		entryURLs[e.Url] = true
	}

	// Spot-check first and last recipe.
	assert.True(t, entryURLs["https://foodblog.example.com/recipes/creamy-tuscan-chicken-pasta"],
		"first recipe article should be discovered")
	assert.True(t, entryURLs["https://foodblog.example.com/recipes/japanese-teriyaki-tofu-bowls"],
		"last recipe article should be discovered")

	// Sidebar category links must not appear (excluded via .sidebar class).
	for _, e := range feed.Entries {
		assert.NotContains(t, e.Url, "/category/", "sidebar category link must be excluded")
	}

	// data-src lazy images must be populated on entries.
	for _, e := range feed.Entries {
		assert.NotEmpty(t, e.Images, "entry %s should have image from data-src attribute", e.Url)
	}

	// Nav and footer links must not appear.
	for _, e := range feed.Entries {
		assert.NotContains(t, e.Url, "/about")
		assert.NotContains(t, e.Url, "/privacy")
	}
}

// TestScrapeFeed_LazyLoadImages verifies that the full ScrapeFeed pipeline
// correctly detects images that use lazy-load attributes (data-src, data-lazy-src,
// srcset, data-lazy) — without any standard "src" attribute — and populates
// entry.Images for each discovered entry.
func TestScrapeFeed_LazyLoadImages(t *testing.T) {
	doc := parseDocumentFile(t, "testdata/lazy_images.html")
	data := &model.DataInput{
		Url:      "https://lazyload.example.com/recipes",
		Document: doc,
	}
	feed := &model.Feed{}

	err := discovery.ScrapeFeed(data, feed, discovery.SamplingOptions{}, discovery.ScoringOptions{})
	require.NoError(t, err)

	assert.Equal(t, discovery.SourceDOMContainer, feed.Discovered.Source)
	require.GreaterOrEqual(t, len(feed.Entries), 10, "all 10 lazy-load recipe entries should be discovered")

	// Every entry must have an image — confirms all lazy-load attr variants are detected.
	for _, e := range feed.Entries {
		assert.NotEmpty(t, e.Images,
			"entry %s must have an image detected from lazy-load attributes", e.Url)
	}
}

// TestScrapeFeed_ReplayUnknownSource verifies that ReplayDiscovered returns an
// error and sets feed.Discovered to nil when the stored Source is unrecognised.
func TestScrapeFeed_ReplayUnknownSource(t *testing.T) {
	discovered := &model.DiscoveredFeed{
		Source:     "unknown-source",
		UrlPattern: "/recipes/",
	}

	const html = `<!DOCTYPE html><html><body><main><ul>
		<li><a href="/recipes/pasta">Pasta</a></li>
	</ul></main></body></html>`

	doc := parseDocument(t, html)
	data := &model.DataInput{Url: "https://example.com", Document: doc}
	feed := &model.Feed{}

	err := discovery.ReplayDiscovered(data, feed, discovered)
	assert.Error(t, err, "unknown source should return an error")
	assert.Nil(t, feed.Discovered, "feed.Discovered should be nil after unknown-source error")
	assert.Empty(t, feed.Entries)
}

// TestScrapeFeed_BulletSeparatedMetadata verifies that links whose visible text
// uses "•" as a metadata separator are correctly scored. textQualityScore must
// split on "•" and evaluate the title segment independently so that a card
// containing "Great Pasta Dish • thirty minutes cooking • beginner friendly" scores
// as "Great Pasta Dish" (3 words → 1.0) rather than the long concatenated string.
func TestScrapeFeed_BulletSeparatedMetadata(t *testing.T) {
	// Each link's full text exceeds 12 words but the title segment is 3–4 words.
	const html = `<!DOCTYPE html><html><body><main><ul>
		<li><a href="/recipes/pasta"><img src="/img/pasta.jpg">Creamy Pasta Bake • thirty minutes cooking • beginner level difficulty rating</a></li>
		<li><a href="/recipes/chicken"><img src="/img/chicken.jpg">Roast Chicken Dinner • one hour total • intermediate skill level needed</a></li>
		<li><a href="/recipes/salad"><img src="/img/salad.jpg">Summer Salad Bowl • ten minutes prep • very easy beginner friendly</a></li>
		<li><a href="/recipes/soup"><img src="/img/soup.jpg">Tomato Basil Soup • twenty five minutes • easy difficulty level rating</a></li>
		<li><a href="/recipes/steak"><img src="/img/steak.jpg">Pan Seared Steak • fifteen minutes only • medium difficulty cooking skill</a></li>
	</ul></main></body></html>`

	doc := parseDocument(t, html)
	data := &model.DataInput{Url: "https://example.com/recipes", Document: doc}
	feed := &model.Feed{}

	err := discovery.ScrapeFeed(data, feed, discovery.SamplingOptions{}, discovery.ScoringOptions{})
	require.NoError(t, err, "bullet-separated metadata should not prevent discovery")
	assert.NotEmpty(t, feed.Entries)
	assert.Equal(t, discovery.SourceDOMContainer, feed.Discovered.Source)
}

// TestScrapeFeed_MinGroupSizeBoundary verifies the minGroupSize=3 boundary: a
// page with exactly 3 qualifying links is accepted, while a page with only 2
// qualifying links is rejected (group filtered before scoring).
func TestScrapeFeed_MinGroupSizeBoundary(t *testing.T) {
	t.Run("exactly three links accepted", func(t *testing.T) {
		const html = `<!DOCTYPE html><html><body><main><ul>
			<li><a href="/recipes/pasta"><img src="/img/pasta.jpg"><span>Creamy Pasta Bake</span></a></li>
			<li><a href="/recipes/chicken"><img src="/img/chicken.jpg"><span>Roast Chicken Dinner</span></a></li>
			<li><a href="/recipes/salad"><img src="/img/salad.jpg"><span>Summer Salad Bowl</span></a></li>
		</ul></main></body></html>`

		doc := parseDocument(t, html)
		data := &model.DataInput{Url: "https://example.com/recipes", Document: doc}
		feed := &model.Feed{}

		err := discovery.ScrapeFeed(data, feed, discovery.SamplingOptions{}, discovery.ScoringOptions{})
		require.NoError(t, err, "group of exactly 3 links should pass minGroupSize filter")
		assert.Len(t, feed.Entries, 3)
	})

	t.Run("two links rejected", func(t *testing.T) {
		const html = `<!DOCTYPE html><html><body><main><ul>
			<li><a href="/recipes/pasta"><img src="/img/pasta.jpg"><span>Creamy Pasta Bake</span></a></li>
			<li><a href="/recipes/chicken"><img src="/img/chicken.jpg"><span>Roast Chicken Dinner</span></a></li>
		</ul></main></body></html>`

		doc := parseDocument(t, html)
		data := &model.DataInput{Url: "https://example.com/recipes", Document: doc}
		feed := &model.Feed{}

		err := discovery.ScrapeFeed(data, feed, discovery.SamplingOptions{}, discovery.ScoringOptions{})
		assert.Error(t, err, "group of 2 links should be filtered out — below minGroupSize=3")
		assert.Empty(t, feed.Entries)
	})
}

// TestScrapeFeed_ScoringOptionsOverride verifies that a non-default ScoringOptions
// actually changes the outcome: a 2-link group that the default minGroupSize (3)
// rejects is accepted once the caller lowers MinGroupSize to 2.
func TestScrapeFeed_ScoringOptionsOverride(t *testing.T) {
	const html = `<!DOCTYPE html><html><body><main><ul>
		<li><a href="/recipes/pasta"><img src="/img/pasta.jpg"><span>Creamy Pasta Bake</span></a></li>
		<li><a href="/recipes/chicken"><img src="/img/chicken.jpg"><span>Roast Chicken Dinner</span></a></li>
	</ul></main></body></html>`

	t.Run("rejected with default thresholds", func(t *testing.T) {
		doc := parseDocument(t, html)
		data := &model.DataInput{Url: "https://example.com/recipes", Document: doc}
		feed := &model.Feed{}

		err := discovery.ScrapeFeed(data, feed, discovery.SamplingOptions{}, discovery.ScoringOptions{})
		assert.Error(t, err, "2-link group is below the default minGroupSize of 3")
		assert.Empty(t, feed.Entries)
	})

	t.Run("accepted with MinGroupSize override", func(t *testing.T) {
		doc := parseDocument(t, html)
		data := &model.DataInput{Url: "https://example.com/recipes", Document: doc}
		feed := &model.Feed{}

		err := discovery.ScrapeFeed(data, feed, discovery.SamplingOptions{}, discovery.ScoringOptions{MinGroupSize: 2})
		require.NoError(t, err, "lowering MinGroupSize to 2 should let the 2-link group through")
		assert.Len(t, feed.Entries, 2)
		assert.Equal(t, discovery.SourceDOMContainer, feed.Discovered.Source)
	})
}

// TestScrapeFeed_RootLevelSlugs verifies that the discovery pipeline handles sites
// (e.g. klopotenko.com) where:
//  1. Recipe URLs are root-level slugs (/recipe-slug/) with no common path prefix.
//  2. Each recipe card has two <a> tags pointing to the same URL: one wraps only an
//     image (data-src lazy loading), the other wraps only the title text.
//  3. A category link (<a href="/category/X/">) lives inside each card and must NOT
//     appear in the discovered entries.
//
// Without fixes, urlConsistencyScore=0 (no common prefix) and the duplicate-link
// pattern halves imageDensityScore and textQualityScore, dropping the total below
// the acceptance threshold. Additionally, "first-seen-wins" dedup in entry building
// causes names to be lost (image link seen first, title link skipped).
func TestScrapeFeed_RootLevelSlugs(t *testing.T) {
	const html = `<!DOCTYPE html><html><body>
	<nav><a href="/about">About</a><a href="/contact">Contact</a></nav>
	<main>
		<section class="recipes-section">
			<div class="row">
				<div class="col"><article class="recipe-card">
					<a href="/creamy-pasta-bake/"><img data-src="/img/pasta.jpg" alt="pasta"></a>
					<div class="recipe-category"><a href="/category/pasta/">Pasta</a></div>
					<a href="/creamy-pasta-bake/"><h3>Creamy Pasta Bake</h3></a>
				</article></div>
				<div class="col"><article class="recipe-card">
					<a href="/roast-chicken-dinner/"><img data-src="/img/chicken.jpg" alt="chicken"></a>
					<div class="recipe-category"><a href="/category/mains/">Mains</a></div>
					<a href="/roast-chicken-dinner/"><h3>Roast Chicken Dinner</h3></a>
				</article></div>
				<div class="col"><article class="recipe-card">
					<a href="/summer-salad-bowl/"><img data-src="/img/salad.jpg" alt="salad"></a>
					<div class="recipe-category"><a href="/category/salads/">Salads</a></div>
					<a href="/summer-salad-bowl/"><h3>Summer Salad Bowl</h3></a>
				</article></div>
				<div class="col"><article class="recipe-card">
					<a href="/tomato-basil-soup/"><img data-src="/img/soup.jpg" alt="soup"></a>
					<div class="recipe-category"><a href="/category/soups/">Soups</a></div>
					<a href="/tomato-basil-soup/"><h3>Tomato Basil Soup</h3></a>
				</article></div>
				<div class="col"><article class="recipe-card">
					<a href="/pan-seared-steak/"><img data-src="/img/steak.jpg" alt="steak"></a>
					<div class="recipe-category"><a href="/category/mains/">Mains</a></div>
					<a href="/pan-seared-steak/"><h3>Pan Seared Steak</h3></a>
				</article></div>
				<div class="col"><article class="recipe-card">
					<a href="/lemon-drizzle-cake/"><img data-src="/img/cake.jpg" alt="cake"></a>
					<div class="recipe-category"><a href="/category/desserts/">Desserts</a></div>
					<a href="/lemon-drizzle-cake/"><h3>Lemon Drizzle Cake</h3></a>
				</article></div>
			</div>
		</section>
	</main>
	<footer><a href="/privacy">Privacy</a><a href="/terms">Terms</a></footer>
	</body></html>`

	doc := parseDocument(t, html)
	data := &model.DataInput{Url: "https://example.com/", Document: doc}
	feed := &model.Feed{}

	err := discovery.ScrapeFeed(data, feed, discovery.SamplingOptions{}, discovery.ScoringOptions{})
	require.NoError(t, err, "ScrapeFeed should succeed on root-level slug layout")

	assert.Equal(t, discovery.SourceDOMContainer, feed.Discovered.Source)
	assert.GreaterOrEqual(t, feed.Discovered.ConfidenceScore, 0.55)

	// Exactly 6 unique recipe entries — NOT 12 (duplicate links must be collapsed).
	require.Len(t, feed.Entries, 6)

	entryURLs := make(map[string]bool, len(feed.Entries))
	for _, e := range feed.Entries {
		entryURLs[e.Url] = true
	}
	assert.True(t, entryURLs["https://example.com/creamy-pasta-bake/"], "first recipe should be discovered")
	assert.True(t, entryURLs["https://example.com/roast-chicken-dinner/"], "second recipe should be discovered")
	assert.True(t, entryURLs["https://example.com/lemon-drizzle-cake/"], "last recipe should be discovered")

	// Category links must NOT appear in results.
	for _, e := range feed.Entries {
		assert.NotContains(t, e.Url, "/category/", "category link must not appear in feed entries")
	}

	// Each entry must have both an image (from data-src) and a name (from title link).
	for _, e := range feed.Entries {
		assert.NotEmpty(t, e.Images, "entry %s should have image from data-src", e.Url)
		assert.NotEmpty(t, e.Name, "entry %s should have name from title link", e.Url)
	}
}

// TestScrapeFeed_ValidationAcceptsMajority ensures majority-confirmed groups are accepted.
func TestScrapeFeed_ValidationAcceptsMajority(t *testing.T) {
	const html = `<!DOCTYPE html><html><body><main><ul>
		<li><a href="/recipes/pasta"><img src="/img/pasta.jpg"><span>Creamy Pasta Bake</span></a></li>
		<li><a href="/recipes/chicken"><img src="/img/chicken.jpg"><span>Roast Chicken Dinner</span></a></li>
		<li><a href="/recipes/salad"><img src="/img/salad.jpg"><span>Summer Salad Bowl</span></a></li>
		<li><a href="/recipes/soup"><img src="/img/soup.jpg"><span>Tomato Basil Soup</span></a></li>
		<li><a href="/recipes/steak"><img src="/img/steak.jpg"><span>Pan-Seared Steak</span></a></li>
	</ul></main></body></html>`

	doc := parseDocument(t, html)
	data := &model.DataInput{Url: "https://example.com/recipes", Document: doc}
	feed := &model.Feed{}

	// Validator confirms all sampled URLs (100% hit rate).
	calls := 0
	validator := func(urls []string) []*model.Recipe {
		calls++
		out := make([]*model.Recipe, len(urls))
		for i, u := range urls {
			out[i] = &model.Recipe{Url: u}
		}
		return out
	}

	err := discovery.ScrapeFeed(data, feed, discovery.SamplingOptions{Validator: validator, SampleSize: 2}, discovery.ScoringOptions{})
	require.NoError(t, err)
	assert.Equal(t, 1, calls, "validator should be called exactly once for the winning group")
	assert.NotEmpty(t, feed.Entries)
	assert.Equal(t, discovery.SourceDOMContainer, feed.Discovered.Source)
}

// TestScrapeFeed_ValidationFallback verifies that when a validator confirms fewer
// than half the sampled URLs, the group is not immediately accepted, but the best-
// effort fallback commits the group with the most confirmed URLs.
func TestScrapeFeed_ValidationFallback(t *testing.T) {
	const html = `<!DOCTYPE html><html><body><main><ul>
		<li><a href="/recipes/pasta"><img src="/img/pasta.jpg"><span>Creamy Pasta Bake</span></a></li>
		<li><a href="/recipes/chicken"><img src="/img/chicken.jpg"><span>Roast Chicken</span></a></li>
		<li><a href="/recipes/salad"><img src="/img/salad.jpg"><span>Summer Salad</span></a></li>
		<li><a href="/recipes/soup"><img src="/img/soup.jpg"><span>Tomato Soup</span></a></li>
		<li><a href="/recipes/steak"><img src="/img/steak.jpg"><span>Pan Steak</span></a></li>
	</ul></main></body></html>`

	doc := parseDocument(t, html)
	data := &model.DataInput{Url: "https://example.com/recipes", Document: doc}
	feed := &model.Feed{}

	// Validator confirms only 1 of 3 sampled URLs — minority, but > 0.
	validator := func(urls []string) []*model.Recipe {
		if len(urls) == 0 {
			return nil
		}
		return []*model.Recipe{{Url: urls[0]}}
	}

	err := discovery.ScrapeFeed(data, feed, discovery.SamplingOptions{Validator: validator, SampleSize: 3}, discovery.ScoringOptions{})
	require.NoError(t, err, "best-effort fallback should succeed when at least one URL is confirmed")
	assert.NotEmpty(t, feed.Entries)
	assert.Equal(t, discovery.SourceDOMContainer, feed.Discovered.Source)
}

// TestScrapeFeed_ValidationRejectsAllGroups ensures failure when no groups are confirmed.
func TestScrapeFeed_ValidationRejectsAllGroups(t *testing.T) {
	const html = `<!DOCTYPE html><html><body><main><ul>
		<li><a href="/recipes/pasta"><img src="/img/pasta.jpg"><span>Creamy Pasta Bake</span></a></li>
		<li><a href="/recipes/chicken"><img src="/img/chicken.jpg"><span>Roast Chicken</span></a></li>
		<li><a href="/recipes/salad"><img src="/img/salad.jpg"><span>Summer Salad</span></a></li>
		<li><a href="/recipes/soup"><img src="/img/soup.jpg"><span>Tomato Soup</span></a></li>
		<li><a href="/recipes/steak"><img src="/img/steak.jpg"><span>Pan Steak</span></a></li>
	</ul></main></body></html>`

	doc := parseDocument(t, html)
	data := &model.DataInput{Url: "https://example.com/recipes", Document: doc}
	feed := &model.Feed{}

	validator := func(urls []string) []*model.Recipe { return nil }

	err := discovery.ScrapeFeed(data, feed, discovery.SamplingOptions{Validator: validator, SampleSize: 3}, discovery.ScoringOptions{})
	assert.Error(t, err, "should fail when no group has any confirmed URLs")
	assert.Empty(t, feed.Entries)
}

// TestScrapeFeed_GridTilePattern verifies that the sibling-merge logic correctly
// collapses recipe tiles that are deeply nested as sibling <div> children of a
// grid container.
//
//  1. Each recipe tile is: grid > tile:nth(N) > tile-inner:nth(1) > <a>.
//     At keyDepth=4, the tiles each form their own group (unique nth-of-type(N)).
//     mergeSiblingGroups must strip TWO levels to find the shared grid prefix.
//
//  2. Each <a> contains a <style> element injected by the SSR framework.
//     linkVisibleText() must strip it so the text scorer sees only human text.
//
//  3. The visible text is "Title | XX Min | Difficulty | YYY kcal".
//     The segment-based textQualityScore must evaluate each "|"-delimited
//     segment and score the title part (3–8 words → 1.0).
//
// The page also has a category carousel (same sibling-div pattern but no images,
// single-word links) which forms a competing group that must NOT win.
func TestScrapeFeed_GridTilePattern(t *testing.T) {
	doc := parseDocumentFile(t, "testdata/grid_tiles.html")
	data := &model.DataInput{
		Url:      "https://cookbook.example.com/recipes/fusion",
		Document: doc,
	}
	feed := &model.Feed{}

	err := discovery.ScrapeFeed(data, feed, discovery.SamplingOptions{}, discovery.ScoringOptions{})
	require.NoError(t, err, "ScrapeFeed should succeed on the grid-tile page")

	assert.Equal(t, discovery.SourceDOMContainer, feed.Discovered.Source,
		"should be discovered via DOM container scoring, not sitemap or RSS")
	assert.GreaterOrEqual(t, feed.Discovered.ConfidenceScore, 0.55,
		"confidence score should clear the acceptance threshold")

	// All 12 individual recipe tiles must be present.
	require.GreaterOrEqual(t, len(feed.Entries), 12,
		"all recipe tiles should be discovered")

	// Build URL set for assertions.
	entryURLs := make(map[string]bool, len(feed.Entries))
	for _, e := range feed.Entries {
		entryURLs[e.Url] = true
	}

	// Spot-check two specific recipe tiles from opposite ends of the grid.
	assert.True(t,
		entryURLs["https://cookbook.example.com/recipes/spaghetti-carbonara-with-pancetta-5a1b2c3d4e5f"],
		"first recipe tile should be discovered")
	assert.True(t,
		entryURLs["https://cookbook.example.com/recipes/japanese-miso-ramen-with-soft-egg-gl2a3b4c5d6e"],
		"last recipe tile should be discovered")

	// Category carousel links (short single-word slugs) must not dominate.
	categoryLinks := 0
	for _, e := range feed.Entries {
		if strings.HasSuffix(e.Url, "/recipes/italian") ||
			strings.HasSuffix(e.Url, "/recipes/asian") ||
			strings.HasSuffix(e.Url, "/recipes/mexican") {
			categoryLinks++
		}
	}
	assert.Zero(t, categoryLinks,
		"category carousel links should not be in the winning recipe group")

	// Nav and footer links must not appear at all.
	for _, e := range feed.Entries {
		assert.NotContains(t, e.Url, "/about", "nav link must be excluded")
		assert.NotContains(t, e.Url, "/privacy", "footer link must be excluded")
	}
}

// TestScrapeFeed_SmallRecipeCardsInSection verifies that a <section> containing
// <article> cards spread across multiple rows and column widths is discovered as
// a single group. This models the klopotenko.com "new-recipes" layout where:
//   - Row 1: two <article> cards share a <div.col-sm-4>, plus one featured card in <div.col-sm-8>
//   - Row 2: three <article> cards each in their own <div.col-sm-4>
//
// Without the article-skip logic, these five cards form three separate sub-groups
// that cannot merge (only 2 child groups at the section level). With article-skip,
// effectiveParent walks past each <article> and its col/row <div> wrappers to land
// directly on the <section>, placing all five entries in one group.
func TestScrapeFeed_SmallRecipeCardsInSection(t *testing.T) {
	const html = `<!DOCTYPE html><html><body>
	<nav><a href="/about">About</a></nav>
	<main>
		<section class="new-recipes">
			<div class="row">
				<div class="col-sm-4">
					<article class="small-recipe-card">
						<a href="/cheesy-pasta-bake/"><img data-src="/img/1.jpg" alt="pasta"></a>
						<a href="/cheesy-pasta-bake/"><h3>Cheesy Pasta Bake</h3></a>
					</article>
					<article class="small-recipe-card">
						<a href="/roast-chicken-dinner/"><img data-src="/img/2.jpg" alt="chicken"></a>
						<a href="/roast-chicken-dinner/"><h3>Roast Chicken Dinner</h3></a>
					</article>
				</div>
				<div class="col-sm-8">
					<article class="recipe-card featured">
						<a href="/summer-salad-bowl/"><img src="/img/3.jpg" alt="salad"></a>
						<a href="/summer-salad-bowl/"><h2>Summer Salad Bowl</h2></a>
					</article>
				</div>
			</div>
			<div class="row mt30">
				<div class="col-sm-4">
					<article class="small-recipe-card">
						<a href="/tomato-basil-soup/"><img data-src="/img/4.jpg" alt="soup"></a>
						<a href="/tomato-basil-soup/"><h3>Tomato Basil Soup</h3></a>
					</article>
				</div>
				<div class="col-sm-4">
					<article class="small-recipe-card">
						<a href="/garlic-butter-shrimp/"><img data-src="/img/5.jpg" alt="shrimp"></a>
						<a href="/garlic-butter-shrimp/"><h3>Garlic Butter Shrimp</h3></a>
					</article>
				</div>
				<div class="col-sm-4">
					<article class="small-recipe-card">
						<a href="/lemon-herb-risotto/"><img data-src="/img/6.jpg" alt="risotto"></a>
						<a href="/lemon-herb-risotto/"><h3>Lemon Herb Risotto</h3></a>
					</article>
				</div>
			</div>
		</section>
		<section class="ingredients-month">
			<div class="row">
				<div class="col"><a href="/ingredient/asparagus/">Asparagus</a></div>
				<div class="col"><a href="/ingredient/spinach/">Spinach</a></div>
				<div class="col"><a href="/ingredient/radish/">Radish</a></div>
				<div class="col"><a href="/ingredient/cabbage/">Cabbage</a></div>
				<div class="col"><a href="/ingredient/quince/">Quince</a></div>
				<div class="col"><a href="/ingredient/daikon/">Daikon</a></div>
				<div class="col"><a href="/ingredient/broccoli/">Broccoli</a></div>
				<div class="col"><a href="/ingredient/squid/">Squid</a></div>
			</div>
		</section>
	</main>
	</body></html>`

	doc := parseDocument(t, html)
	data := &model.DataInput{Url: "https://example.com/", Document: doc}
	feed := &model.Feed{}

	err := discovery.ScrapeFeed(data, feed, discovery.SamplingOptions{}, discovery.ScoringOptions{})
	require.NoError(t, err, "ScrapeFeed should discover the recipe section, not fail")

	assert.Equal(t, discovery.SourceDOMContainer, feed.Discovered.Source)
	assert.GreaterOrEqual(t, feed.Discovered.ConfidenceScore, 0.55)

	// All 6 unique recipe URLs must be discovered.
	require.Len(t, feed.Entries, 6, "all recipe entries from both rows should be collected under the section")

	entryURLs := make(map[string]bool, len(feed.Entries))
	for _, e := range feed.Entries {
		entryURLs[e.Url] = true
	}
	assert.True(t, entryURLs["https://example.com/cheesy-pasta-bake/"], "first card (row-1 shared col) should be discovered")
	assert.True(t, entryURLs["https://example.com/lemon-herb-risotto/"], "last card (row-2) should be discovered")

	// Ingredient links must not appear in results.
	for _, e := range feed.Entries {
		assert.NotContains(t, e.Url, "/ingredient/", "ingredient links must not leak into recipe results")
		assert.NotEmpty(t, e.Images, "entry %s should have an image", e.Url)
		assert.NotEmpty(t, e.Name, "entry %s should have a name", e.Url)
	}
}

// TestScrapeFeed_IngredientSectionNoise verifies that a page containing two
// separate groups of recipe-cards (root-level slugs, long titles, article tags,
// split across two parent divs) alongside a competing "ingredients of the month"
// section (common /ingredient/ prefix, short 1-word titles, div containers) ranks
// a recipe-card group first. This reproduces the klopotenko.com pattern where the
// recipe grid is split into sub-groups of 4 and the /ingredient/ section would
// otherwise win on URL-consistency alone.
func TestScrapeFeed_IngredientSectionNoise(t *testing.T) {
	// Two groups of 4 recipe-cards that cannot merge with each other: each group
	// sits under a distinct top-level <div> inside the recipes section, so their
	// container keys share no common ancestor prefix within keyDepth=4.
	// The ingredient section has 8 entries with a /ingredient/ common prefix.
	const html = `<!DOCTYPE html><html><body>
	<nav><a href="/about">About</a></nav>
	<main>
		<section class="new-recipes">
			<h2>New Recipes</h2>
			<div class="featured-col">
				<div class="grid">
					<div class="col"><article class="recipe-card">
						<a href="/cheesy-baked-pasta/"><img data-src="/img/pasta.jpg" alt="pasta"></a>
						<div class="recipe-category"><a href="/category/pasta/">Pasta</a></div>
						<a href="/cheesy-baked-pasta/"><h3>Cheesy Baked Pasta</h3></a>
					</article></div>
					<div class="col"><article class="recipe-card">
						<a href="/herb-roast-chicken/"><img data-src="/img/chicken.jpg" alt="chicken"></a>
						<div class="recipe-category"><a href="/category/mains/">Mains</a></div>
						<a href="/herb-roast-chicken/"><h3>Herb Roast Chicken</h3></a>
					</article></div>
					<div class="col"><article class="recipe-card">
						<a href="/spring-vegetable-soup/"><img data-src="/img/soup.jpg" alt="soup"></a>
						<div class="recipe-category"><a href="/category/soups/">Soups</a></div>
						<a href="/spring-vegetable-soup/"><h3>Spring Vegetable Soup</h3></a>
					</article></div>
					<div class="col"><article class="recipe-card">
						<a href="/lemon-poppy-seed-cake/"><img data-src="/img/cake.jpg" alt="cake"></a>
						<div class="recipe-category"><a href="/category/desserts/">Desserts</a></div>
						<a href="/lemon-poppy-seed-cake/"><h3>Lemon Poppy Seed Cake</h3></a>
					</article></div>
				</div>
			</div>
			<div class="listing-col">
				<div class="grid">
					<div class="col"><article class="recipe-card">
						<a href="/grilled-halloumi-salad/"><img data-src="/img/salad.jpg" alt="salad"></a>
						<div class="recipe-category"><a href="/category/salads/">Salads</a></div>
						<a href="/grilled-halloumi-salad/"><h3>Grilled Halloumi Salad</h3></a>
					</article></div>
					<div class="col"><article class="recipe-card">
						<a href="/slow-cooker-beef-stew/"><img data-src="/img/stew.jpg" alt="stew"></a>
						<div class="recipe-category"><a href="/category/mains/">Mains</a></div>
						<a href="/slow-cooker-beef-stew/"><h3>Slow Cooker Beef Stew</h3></a>
					</article></div>
					<div class="col"><article class="recipe-card">
						<a href="/mango-coconut-pudding/"><img data-src="/img/pudding.jpg" alt="pudding"></a>
						<div class="recipe-category"><a href="/category/desserts/">Desserts</a></div>
						<a href="/mango-coconut-pudding/"><h3>Mango Coconut Pudding</h3></a>
					</article></div>
					<div class="col"><article class="recipe-card">
						<a href="/crispy-fish-tacos/"><img data-src="/img/tacos.jpg" alt="tacos"></a>
						<div class="recipe-category"><a href="/category/mains/">Mains</a></div>
						<a href="/crispy-fish-tacos/"><h3>Crispy Fish Tacos</h3></a>
					</article></div>
				</div>
			</div>
		</section>
		<section class="ingredients-month">
			<h2>Ingredients of the Month</h2>
			<div class="row">
				<div class="col"><a href="/ingredient/radish/" class="ingredient-link"><div><img data-src="/img/radish.jpg"><h3>Radish</h3></div></a></div>
				<div class="col"><a href="/ingredient/asparagus/" class="ingredient-link"><div><img data-src="/img/asparagus.jpg"><h3>Asparagus</h3></div></a></div>
				<div class="col"><a href="/ingredient/spinach/" class="ingredient-link"><div><img data-src="/img/spinach.jpg"><h3>Spinach</h3></div></a></div>
				<div class="col"><a href="/ingredient/cabbage/" class="ingredient-link"><div><img data-src="/img/cabbage.jpg"><h3>Cabbage</h3></div></a></div>
				<div class="col"><a href="/ingredient/quince/" class="ingredient-link"><div><img data-src="/img/quince.jpg"><h3>Quince</h3></div></a></div>
				<div class="col"><a href="/ingredient/daikon/" class="ingredient-link"><div><img data-src="/img/daikon.jpg"><h3>Daikon</h3></div></a></div>
				<div class="col"><a href="/ingredient/broccoli/" class="ingredient-link"><div><img data-src="/img/broccoli.jpg"><h3>Broccoli</h3></div></a></div>
				<div class="col"><a href="/ingredient/squid/" class="ingredient-link"><div><img data-src="/img/squid.jpg"><h3>Squid</h3></div></a></div>
			</div>
		</section>
	</main>
	<footer><a href="/privacy">Privacy</a></footer>
	</body></html>`

	doc := parseDocument(t, html)
	data := &model.DataInput{Url: "https://example.com/", Document: doc}
	feed := &model.Feed{}

	err := discovery.ScrapeFeed(data, feed, discovery.SamplingOptions{}, discovery.ScoringOptions{})
	require.NoError(t, err, "ScrapeFeed should succeed and pick a recipe-card group, not ingredients")

	assert.Equal(t, discovery.SourceDOMContainer, feed.Discovered.Source)
	assert.GreaterOrEqual(t, feed.Discovered.ConfidenceScore, 0.55)

	for _, e := range feed.Entries {
		assert.NotContains(t, e.Url, "/ingredient/", "ingredient links must not appear in results")
		assert.NotContains(t, e.Url, "/category/", "category links must not appear in results")
		assert.NotEmpty(t, e.Images, "entry %s should have an image from data-src", e.Url)
		assert.NotEmpty(t, e.Name, "entry %s should have a name from the title link", e.Url)
	}
}
