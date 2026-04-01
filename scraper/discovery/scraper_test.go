package discovery_test

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/scraper/discovery"
)

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

	err := discovery.ScrapeFeed(data, feed)
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

	err := discovery.ScrapeFeed(data, feed)
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
// from the DOM container during initial discovery (SkipEntriesScrape = true scenario).
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

	err := discovery.ScrapeFeed(data, feed)
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

	err := discovery.ScrapeFeed(data, feed)
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

	err := discovery.ScrapeFeed(data, feed)
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
		err := discovery.ScrapeFeed(data, feed)
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
		err := discovery.ScrapeFeed(data, feed)
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

	err := discovery.ScrapeFeed(data, feed)
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
