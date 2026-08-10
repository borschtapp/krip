package scraper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/scraper/custom"
)

func makeRecipe(name, rawUrl string, withImage, withIngredients, withInstructions bool) *model.Recipe {
	r := &model.Recipe{Name: name, Url: rawUrl, Publisher: &model.Organization{Name: "Pub"}}
	if withImage {
		r.Images = []*model.ImageObject{{Url: "img.jpg"}}
	}
	if withIngredients {
		r.Ingredients = []*model.PropertyValue{{Name: "ingredient"}}
	}
	if withInstructions {
		r.Instructions = []*model.HowToSection{{HowToStep: model.HowToStep{Text: "cook"}}}
	}
	return r
}

func TestFilterEntries_RequiresImageByDefault(t *testing.T) {
	entries := []*model.Recipe{
		makeRecipe("With Image", "https://x.com/r/1", true, true, true),
		makeRecipe("No Image", "https://x.com/r/2", false, true, true),
	}

	entries[0].Scraped = true
	entries[1].Scraped = true

	result := filterEntries(entries, model.FeedOptions{})

	assert.Len(t, result, 1)
	assert.Equal(t, "With Image", result[0].Name)
}

func TestFilterEntries_OptionalImageAcceptsAll(t *testing.T) {
	entries := []*model.Recipe{
		makeRecipe("A", "https://x.com/r/1", false, true, true),
		makeRecipe("B", "https://x.com/r/2", false, true, true),
	}

	entries[0].Scraped = true
	entries[1].Scraped = true

	opt := model.FeedOptions{RecipeFilter: model.RecipeFilter{OptionalImage: true}}
	result := filterEntries(entries, opt)

	assert.Len(t, result, 2)
}

func TestFilterEntries_EmptyUrlNotExemptFromValidation(t *testing.T) {
	noUrl := makeRecipe("No URL", "", true, true, true)
	noUrl.Scraped = true

	deferredStub := makeRecipe("Deferred stub", "https://x.com/r/2", false, false, false) // never attempted, Scraped stays false

	scrapedIncomplete := makeRecipe("Scraped but incomplete", "https://x.com/r/1", false, true, true)
	scrapedIncomplete.Scraped = true

	entries := []*model.Recipe{noUrl, deferredStub, scrapedIncomplete}

	result := filterEntries(entries, model.FeedOptions{})

	var names []string
	for _, r := range result {
		names = append(names, r.Name)
	}
	assert.NotContains(t, names, "No URL", "entry with empty Url must not bypass Validate()")
	assert.Contains(t, names, "Deferred stub", "unscraped stub entries should still be exempt from validation")
	assert.NotContains(t, names, "Scraped but incomplete", "attempted-but-incomplete entries must still be validated")
}

// TestFilterEntries_SurvivesUrlRewrite is a regression test for the bug where Scrape
// rewrites entry.Url (redirects, JSON-LD canonical url) after the caller already
// recorded which entries it attempted. Tracking by a Scraped flag on the entry itself
// - set by Scrape() - rather than by URL means the mutation can no longer defeat
// RecipeFilter.
func TestFilterEntries_SurvivesUrlRewrite(t *testing.T) {
	entry := makeRecipe("Trailing Slash", "https://x.com/r/1", false, false, false) // no image/ingredients/instructions -> should fail default filter

	dataInput := &model.DataInput{Url: "https://x.com/r/1/"} // simulates a redirect adding a trailing slash
	err := Scrape(dataInput, entry, model.ScrapeOptions{SkipSchemaScraper: true, SkipOpenGraphScraper: true, SkipCustomScrapers: true})
	require.NoError(t, err)
	require.Equal(t, "https://x.com/r/1/", entry.Url, "Scrape must have rewritten the URL for this test to be meaningful")

	result := filterEntries([]*model.Recipe{entry}, model.FeedOptions{})

	assert.Empty(t, result, "a scraped entry that fails RecipeFilter must be discarded even though its Url changed mid-scrape")
}

func TestFilterEntries_MinIngredients(t *testing.T) {
	few := makeRecipe("Few", "https://x.com/r/1", true, false, true)
	few.Ingredients = []*model.PropertyValue{{Name: "salt"}}

	enough := makeRecipe("Enough", "https://x.com/r/2", true, false, true)
	enough.Ingredients = []*model.PropertyValue{{Name: "a"}, {Name: "b"}, {Name: "c"}}

	few.Scraped = true
	enough.Scraped = true

	opt := model.FeedOptions{RecipeFilter: model.RecipeFilter{MinIngredients: 3}}
	result := filterEntries([]*model.Recipe{few, enough}, opt)

	assert.Len(t, result, 1)
	assert.Equal(t, "Enough", result[0].Name)
}

func TestValidateDiscoverySample_StrictMajority(t *testing.T) {
	// hits*2 > tried: for tried=2, both must succeed (2*2=4 > 2)
	// for tried=3, two must succeed (2*2=4 > 3)
	tests := []struct {
		tried int
		hits  int
		want  bool
	}{
		{1, 1, true},
		{1, 0, false},
		{2, 2, true},  // both succeed
		{2, 1, false}, // only one — NOT a majority for tried=2
		{3, 2, true},  // 2/3 — majority
		{3, 1, false}, // 1/3 — not majority
	}

	for _, tc := range tests {
		// Check majority.
		got := tc.tried > 0 && tc.hits*2 > tc.tried
		assert.Equal(t, tc.want, got, "tried=%d hits=%d", tc.tried, tc.hits)
	}
}

func TestFindEntries_SingleInvalidSchemaEntryFallsThrough(t *testing.T) {
	const rawHTML = `<!DOCTYPE html><html><head>
<script type="application/ld+json">
{"@context":"https://schema.org","@type":"Recipe","name":"Carbonara","url":"https://example.com/pasta"}
</script>
</head><body><main><ul>
  <li><a href="/recipes/pasta"><img src="/img/pasta.jpg"><span>Creamy Pasta</span></a></li>
  <li><a href="/recipes/chicken"><img src="/img/chicken.jpg"><span>Roast Chicken</span></a></li>
  <li><a href="/recipes/salad"><img src="/img/salad.jpg"><span>Summer Salad</span></a></li>
  <li><a href="/recipes/soup"><img src="/img/soup.jpg"><span>Tomato Soup</span></a></li>
</ul></main></body></html>`

	root, err := html.Parse(strings.NewReader(rawHTML))
	require.NoError(t, err)
	data, err := NodeInput(root, "https://example.com/recipes", model.ScrapeOptions{SkipMetaUrl: true})
	require.NoError(t, err)

	feed := &model.Feed{}
	err = findEntries(data, feed, model.FeedOptions{})
	require.NoError(t, err, "should succeed by falling through to discovery")

	// Discovery must find the <ul> entries, not the schema stub.
	assert.Greater(t, len(feed.Entries), 1, "should find multiple entries via DOM discovery, not just the schema stub")
	for _, e := range feed.Entries {
		assert.NotEqual(t, "https://example.com/pasta", e.Url, "schema stub URL must not appear")
	}
}

type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

func TestScrapeFeed_ContextCancelledEntryScrape(t *testing.T) {
	// Register a dummy custom feed scraper to yield some entries
	custom.RegisterFeedScraper("canceltestentry", func(data *model.DataInput, feed *model.Feed) error {
		feed.Entries = []*model.Recipe{
			{Url: "https://canceltestentry.com/1"},
			{Url: "https://canceltestentry.com/2"},
			{Url: "https://canceltestentry.com/3"},
		}
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	callCount := 0
	mockClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			return nil, context.Canceled
		},
	}

	opt := model.FeedOptions{
		ScrapeOptions: model.ScrapeOptions{
			RequestOptions: model.RequestOptions{
				Context:    ctx,
				HttpClient: mockClient,
			},
		},
		SkipFeedMeta: true,
	}

	data := &model.DataInput{
		Url: "https://canceltestentry.com/feed",
	}
	feed := &model.Feed{}

	err := ScrapeFeed(data, feed, opt)
	require.NoError(t, err) // ScrapeFeed returns nil on success or non-critical errors

	// The loop should break immediately on the cancelled context, so Do should be called 0 times
	assert.Equal(t, 0, callCount, "should not make any requests if context is cancelled")
}

func TestScrapeFeed_ContextCancelledDiscoveryValidation(t *testing.T) {
	// Construct HTML with multiple candidate links
	const rawHTML = `<!DOCTYPE html><html><body><main><ul>
	  <li><a href="/recipes/pasta"><img src="/img/pasta.jpg"><span>Creamy Pasta</span></a></li>
	  <li><a href="/recipes/chicken"><img src="/img/chicken.jpg"><span>Roast Chicken</span></a></li>
	  <li><a href="/recipes/salad"><img src="/img/salad.jpg"><span>Summer Salad</span></a></li>
	  <li><a href="/recipes/soup"><img src="/img/soup.jpg"><span>Tomato Soup</span></a></li>
	</ul></main></body></html>`

	root, err := html.Parse(strings.NewReader(rawHTML))
	require.NoError(t, err)
	data, err := NodeInput(root, "https://example.com/recipes", model.ScrapeOptions{SkipMetaUrl: true})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	callCount := 0
	mockClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			return nil, context.Canceled
		},
	}

	opt := model.FeedOptions{
		ScrapeOptions: model.ScrapeOptions{
			RequestOptions: model.RequestOptions{
				Context:    ctx,
				HttpClient: mockClient,
			},
			SkipCustomScrapers: true,
			SkipSchemaScraper:  true,
		},
		SkipFeedMeta:         true,
		SkipRSSScraper:       true,
		SkipDiscoveryScraper: false,
		DiscoverySampleSize:  3,
		SkipEntriesScrape:    true, // Skip the entry scrape step to focus only on discovery sampling
	}

	feed := &model.Feed{}
	err = ScrapeFeed(data, feed, opt)
	// Since all validation fails (and context is cancelled), findEntries will return "no entries found"
	require.Error(t, err)

	// Currently, the validator loop in findEntries (scraper.go) does NOT check options.Context.Err().
	// It will make 3 calls (DiscoverySampleSize) to UrlInput, which calls client.Do (mockClient).
	// If the issue is fixed, it should make 0 calls because it checks context cancellation before starting.
	assert.Equal(t, 0, callCount, "should not make discovery validation requests if context is cancelled")
}

func TestScrapeFeed_RefinementDenominatorUsesAttemptedCount(t *testing.T) {
	custom.RegisterFeedScraper("refine_denom_test", func(data *model.DataInput, feed *model.Feed) error {
		var entries []*model.Recipe
		// 4 valid recipe entries
		for i := 1; i <= 4; i++ {
			entries = append(entries, &model.Recipe{Name: fmt.Sprintf("Recipe %d", i), Url: fmt.Sprintf("https://refine_denom_test.com/recipes/dish-%d", i)})
		}
		// 6 invalid entries (will fail validation when scraped)
		for i := 1; i <= 6; i++ {
			entries = append(entries, &model.Recipe{Name: fmt.Sprintf("About %d", i), Url: fmt.Sprintf("https://refine_denom_test.com/about-%d", i)})
		}
		// 20 unattempted stubs
		for i := 1; i <= 20; i++ {
			entries = append(entries, &model.Recipe{Name: fmt.Sprintf("Stub %d", i), Url: fmt.Sprintf("https://refine_denom_test.com/recipes/stub-%d", i)})
		}
		feed.Entries = entries
		return nil
	})

	mockClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			urlStr := req.URL.String()
			body := `<html><body>No recipe</body></html>`
			if strings.Contains(urlStr, "/recipes/dish-") {
				body = `<!DOCTYPE html><html><head><script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "Recipe",
  "name": "Dish",
  "image": "img.jpg",
  "publisher": {"@type": "Organization", "name": "Pub"},
  "recipeIngredient": ["salt"],
  "recipeInstructions": [{"@type": "HowToStep", "text": "cook"}]
}
</script></head><body></body></html>`
			}
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		},
	}

	opt := model.FeedOptions{
		ScrapeOptions: model.ScrapeOptions{
			RequestOptions: model.RequestOptions{
				HttpClient: mockClient,
			},
		},
		MaxEntriesForScrape: 10,
		SkipFeedMeta:        true,
		Discovered:          &model.DiscoveredFeed{},
	}

	data := &model.DataInput{Url: "https://refine_denom_test.com/feed"}
	feed := &model.Feed{}

	err := ScrapeFeed(data, feed, opt)
	require.NoError(t, err)

	// Out of 10 attempted entries (4 valid dish recipes + 6 invalid about pages), 6 failed validation (60% > 30%).
	// The 20 stubs were unattempted and remained in feed.Entries.
	// Total initialCount = 30. If initialCount were used as denominator, 6/30 = 20% <= 30% (refinement would NOT trigger).
	// With attemptedCount = 10 used as denominator, 6/10 = 60% > 30% (refinement DOES trigger).
	assert.Equal(t, "/recipes/", feed.Discovered.UrlPattern, "refinement should trigger based on attempted entries discard ratio")
}

func TestScrapeFeed_DiscoveredNotMutatedCallerOptions(t *testing.T) {
	callerDiscovered := &model.DiscoveredFeed{
		Source:     "sitemap",
		Selector:   "https://example.com/sitemap.xml",
		UrlPattern: "",
	}

	opts := model.FeedOptions{
		Discovered:           callerDiscovered,
		SkipEntriesScrape:    true,
		SkipRSSScraper:       true,
		SkipDiscoveryScraper: true,
	}

	data := &model.DataInput{Url: "https://example.com/"}
	feed := &model.Feed{}

	_ = ScrapeFeed(data, feed, opts)

	if feed.Discovered != nil {
		feed.Discovered.UrlPattern = "/mutated-pattern/"
	}

	assert.Equal(t, "", callerDiscovered.UrlPattern, "caller options.Discovered struct should remain unmutated")
}
