package scraper

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"

	"github.com/borschtapp/krip/model"
)

// validRecipePage renders a minimal page that passes the default RecipeFilter:
// name/url (schema), images (schema), publisher (opengraph og:site_name),
// ingredients and instructions (schema).
const validRecipePage = `<!DOCTYPE html><html><head>
<meta property="og:site_name" content="Test Kitchen">
<script type="application/ld+json">{"@context":"https://schema.org","@type":"Recipe","name":"%s","image":"https://example.com/img.jpg","recipeIngredient":["a","b","c"],"recipeInstructions":["step one","step two"]}</script>
</head><body></body></html>`

// TestFindEntries_AdaptiveSampling_StopsEarlyOnMajority verifies the discovery
// sample Validator (scraper.go) stops fetching once the accept/reject verdict
// (hits*2 > n) is already decided, instead of always fetching the full sample.
// The DOM group here has exactly 4 links with DiscoverySampleSize=4, so the
// sample is the whole group in document order (1, 2, 3, 4). The first 3 all
// validate successfully: hits*2 (6) > n (4) after the 3rd, so the 4th must never
// be fetched.
func TestFindEntries_AdaptiveSampling_StopsEarlyOnMajority(t *testing.T) {
	var entryRequests int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/recipes/") {
			atomic.AddInt32(&entryRequests, 1)
			_, _ = fmt.Fprintf(w, validRecipePage, "Recipe at "+r.URL.Path) // #nosec G705
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	listingHTML := fmt.Sprintf(`<!DOCTYPE html><html><body><main><ul>
		<li><a href="%[1]s/recipes/1"><img src="/i/1.jpg"><span>Recipe One Alpha</span></a></li>
		<li><a href="%[1]s/recipes/2"><img src="/i/2.jpg"><span>Recipe Two Beta</span></a></li>
		<li><a href="%[1]s/recipes/3"><img src="/i/3.jpg"><span>Recipe Three Gamma</span></a></li>
		<li><a href="%[1]s/recipes/4"><img src="/i/4.jpg"><span>Recipe Four Delta</span></a></li>
	</ul></main></body></html>`, server.URL)

	root, err := html.Parse(strings.NewReader(listingHTML))
	require.NoError(t, err)
	data, err := NodeInput(root, server.URL+"/recipes", model.ScrapeOptions{SkipMetaUrl: true})
	require.NoError(t, err)

	options := model.FeedOptions{
		DiscoverySampleSize: 4,
		ScrapeOptions: model.ScrapeOptions{
			RequestOptions: model.RequestOptions{HttpClient: server.Client()},
		},
	}

	feed := &model.Feed{}
	err = findEntries(data, feed, options)

	require.NoError(t, err)
	assert.Equal(t, int32(3), atomic.LoadInt32(&entryRequests),
		"validator should stop after the 3rd confirmation (hits*2=6 > n=4) and never fetch the 4th sample URL")
	assert.Len(t, feed.Entries, 4, "all 4 group entries should still be in the feed as stubs/confirmed, even though only 3 were fetched")
}

// TestFindEntries_AdaptiveSampling_StopsEarlyOnRejection verifies the same
// early-stop on the reject side: once enough misses accumulate that a majority
// can no longer be reached, remaining sample URLs must not be fetched either.
func TestFindEntries_AdaptiveSampling_StopsEarlyOnRejection(t *testing.T) {
	var entryRequests int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/recipes/") {
			atomic.AddInt32(&entryRequests, 1)
			// None of these pages satisfy the default RecipeFilter (no ingredients,
			// no instructions, no publisher, no image) - every sample is a miss.
			_, _ = w.Write([]byte(`<!DOCTYPE html><html><body>not a recipe</body></html>`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	listingHTML := fmt.Sprintf(`<!DOCTYPE html><html><body><main><ul>
		<li><a href="%[1]s/recipes/1"><img src="/i/1.jpg"><span>Recipe One Alpha</span></a></li>
		<li><a href="%[1]s/recipes/2"><img src="/i/2.jpg"><span>Recipe Two Beta</span></a></li>
		<li><a href="%[1]s/recipes/3"><img src="/i/3.jpg"><span>Recipe Three Gamma</span></a></li>
		<li><a href="%[1]s/recipes/4"><img src="/i/4.jpg"><span>Recipe Four Delta</span></a></li>
	</ul></main></body></html>`, server.URL)

	root, err := html.Parse(strings.NewReader(listingHTML))
	require.NoError(t, err)
	data, err := NodeInput(root, server.URL+"/recipes", model.ScrapeOptions{SkipMetaUrl: true})
	require.NoError(t, err)

	options := model.FeedOptions{
		DiscoverySampleSize: 4,
		ScrapeOptions: model.ScrapeOptions{
			RequestOptions: model.RequestOptions{HttpClient: server.Client()},
		},
	}

	feed := &model.Feed{}
	// This group is expected to fail validation entirely (no group confirms any
	// URL), which is a valid "no entries found" outcome; we only care about the
	// request count spent getting there.
	_ = findEntries(data, feed, options)

	assert.Equal(t, int32(2), atomic.LoadInt32(&entryRequests),
		"after 2 misses out of n=4, a majority (needs hits>=3) is already impossible, so the remaining 2 samples must not be fetched")
}
