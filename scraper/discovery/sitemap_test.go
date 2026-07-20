package discovery

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/borschtapp/krip/model"
)

func TestParseSitemap_Urlset(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url><loc>https://example.com/recipes/pasta</loc></url>
  <url><loc>https://example.com/recipes/chicken</loc></url>
  <url><loc>https://example.com/about</loc></url>
  <url><loc>https://example.com/contact</loc></url>
</urlset>`

	locs, isIndex, err := parseSitemap([]byte(xml))
	require.NoError(t, err)
	assert.False(t, isIndex)
	assert.Len(t, locs, 4)
	assert.Contains(t, locs, "https://example.com/recipes/pasta")
}

func TestParseSitemap_Index(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <sitemap><loc>https://example.com/recipe-sitemap.xml</loc></sitemap>
  <sitemap><loc>https://example.com/page-sitemap.xml</loc></sitemap>
</sitemapindex>`

	locs, isIndex, err := parseSitemap([]byte(xml))
	require.NoError(t, err)
	assert.True(t, isIndex)
	assert.Len(t, locs, 2)
	assert.Contains(t, locs, "https://example.com/recipe-sitemap.xml")
}

func TestParseSitemap_Invalid(t *testing.T) {
	_, _, err := parseSitemap([]byte("<html><body>not a sitemap</body></html>"))
	assert.Error(t, err)
}

func TestFilterRecipeLocs(t *testing.T) {
	locs := []string{
		"https://example.com/recipes/pasta",
		"https://example.com/recipes/chicken",
		"https://example.com/about",
		"https://example.com/contact",
		"https://example.com/cook/risotto",
		"https://example.com/meal/prep-guide",
		"https://example.com/blog/post-1",
	}

	filtered := filterRecipeLocs(locs)
	assert.Len(t, filtered, 4)
	assert.Contains(t, filtered, "https://example.com/recipes/pasta")
	assert.Contains(t, filtered, "https://example.com/recipes/chicken")
	assert.Contains(t, filtered, "https://example.com/cook/risotto")
	assert.Contains(t, filtered, "https://example.com/meal/prep-guide")
	assert.NotContains(t, filtered, "https://example.com/about")
	assert.NotContains(t, filtered, "https://example.com/blog/post-1")
}

func TestFilterRecipeLocs_Dedup(t *testing.T) {
	locs := []string{
		"https://example.com/recipes/pasta",
		"https://example.com/recipes/pasta", // duplicate
		"https://example.com/recipes/chicken",
	}
	filtered := filterRecipeLocs(locs)
	assert.Len(t, filtered, 2)
}

// TestFilterRecipeLocs_HostKeywordNotPath guards against matching the whole URL
// (including hostname). Sites like recipes.example.com or mycookingblog.com would
// otherwise match on every single loc regardless of path, defeating the filter.
func TestFilterRecipeLocs_HostKeywordNotPath(t *testing.T) {
	locs := []string{
		"https://recipes.example.com/about-us",
		"https://recipes.example.com/contact",
		"https://mycookingblog.com/privacy-policy",
		"https://recipes.example.com/recipes/pasta",
	}

	filtered := filterRecipeLocs(locs)
	assert.Equal(t, []string{"https://recipes.example.com/recipes/pasta"}, filtered,
		"only the loc whose PATH contains a recipe keyword should match, not ones matching only via hostname")
}

func TestMatchesRecipePath_PathOnly(t *testing.T) {
	assert.True(t, matchesRecipePath("https://example.com/recipes/pasta"))
	assert.False(t, matchesRecipePath("https://recipes.example.com/about"),
		"hostname keyword must not count as a path match")
	assert.False(t, matchesRecipePath("://not a url"))
}

func TestCommonPathPrefix(t *testing.T) {
	tests := []struct {
		name     string
		locs     []string
		expected string
	}{
		{
			name: "common /recipes/ prefix",
			locs: []string{
				"https://example.com/recipes/pasta",
				"https://example.com/recipes/chicken",
				"https://example.com/recipes/stew",
			},
			expected: "/recipes/",
		},
		{
			name: "no common prefix",
			locs: []string{
				"https://example.com/recipes/pasta",
				"https://example.com/cook/chicken",
			},
			expected: "",
		},
		{
			name:     "empty input",
			locs:     []string{},
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := UrlPathPattern(tc.locs)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestParseSitemap_SinglePass(t *testing.T) {
	// Verify that parseSitemap correctly distinguishes urlset vs sitemapindex
	// with a single decoder pass (not double unmarshal).
	t.Run("urlset with namespace", func(t *testing.T) {
		xmlData := `<?xml version="1.0"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url><loc>https://example.com/recipes/pasta</loc></url>
  <url><loc>https://example.com/recipes/chicken</loc></url>
</urlset>`
		locs, isIndex, err := parseSitemap([]byte(xmlData))
		require.NoError(t, err)
		assert.False(t, isIndex)
		assert.Len(t, locs, 2)
	})

	t.Run("sitemapindex with namespace", func(t *testing.T) {
		xmlData := `<?xml version="1.0"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <sitemap><loc>https://example.com/recipe-sitemap.xml</loc></sitemap>
</sitemapindex>`
		locs, isIndex, err := parseSitemap([]byte(xmlData))
		require.NoError(t, err)
		assert.True(t, isIndex)
		assert.Len(t, locs, 1)
	})

	t.Run("unknown root element", func(t *testing.T) {
		xmlData := `<?xml version="1.0"?><rss version="2.0"><channel></channel></rss>`
		_, _, err := parseSitemap([]byte(xmlData))
		assert.Error(t, err)
	})

	t.Run("empty urlset returns error", func(t *testing.T) {
		xmlData := `<?xml version="1.0"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"></urlset>`
		_, _, err := parseSitemap([]byte(xmlData))
		assert.Error(t, err, "empty urlset should be an error")
	})
}

func TestMaybeDecompress_PlainData(t *testing.T) {
	data := []byte("<xml>not compressed</xml>")
	result, err := maybeDecompress(data)
	require.NoError(t, err)
	assert.Equal(t, data, result)
}

type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

func TestTrySitemap_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	callCount := 0
	mockClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			return nil, context.Canceled
		},
	}

	opts := model.RequestOptions{
		Context:    ctx,
		HttpClient: mockClient,
	}

	data := &model.DataInput{
		Url: "https://example.com",
	}
	feed := &model.Feed{}
	baseUrl, err := url.Parse(data.Url)
	require.NoError(t, err)

	_, err = trySitemap(data, feed, baseUrl, opts)
	assert.Error(t, err)
	assert.Equal(t, 0, callCount, "should not make sitemap candidate requests if context is cancelled")
}

func TestSitemapCandidates_EnforcesSameHost(t *testing.T) {
	baseUrl, err := url.Parse("https://example.com")
	require.NoError(t, err)

	htmlStr := `<html><head><link rel="sitemap" href="https://malicious.com/sitemap.xml"></head></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlStr))
	require.NoError(t, err)

	data := &model.DataInput{
		Url:      "https://example.com",
		Document: doc,
	}

	candidates := sitemapCandidates(data, baseUrl)
	assert.NotContains(t, candidates, "https://malicious.com/sitemap.xml")
	assert.Contains(t, candidates, "https://example.com/sitemap.xml")
	assert.Contains(t, candidates, "https://example.com/sitemap_index.xml")
}

func TestFollowSitemapIndex_BlocksPrivateAndLoopback(t *testing.T) {
	indexLocs := []string{
		"http://127.0.0.1/sitemap.xml",
		"http://localhost/sitemap.xml",
		"http://192.168.1.1/sitemap.xml",
	}
	opts := model.RequestOptions{
		HttpClient: &mockHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				t.Fatalf("should not perform request to private/loopback URL: %s", req.URL.String())
				return nil, nil
			},
		},
	}

	locs, err := followSitemapIndex(indexLocs, opts)
	require.NoError(t, err)
	assert.Empty(t, locs)
}

func TestReplaySitemap_EnforcesSameHost(t *testing.T) {
	data := &model.DataInput{
		Url: "https://example.com",
	}
	feed := &model.Feed{}

	d := &model.DiscoveredFeed{
		Source:   SourceSitemap,
		Selector: "https://malicious.com/sitemap.xml",
	}

	err := replaySitemap(data, feed, d)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "external host")
}

func TestReplayRSS_EnforcesSameHost(t *testing.T) {
	data := &model.DataInput{
		Url: "https://example.com",
	}
	feed := &model.Feed{}

	d := &model.DiscoveredFeed{
		Source:   SourceRSSLink,
		Selector: "https://malicious.com/feed.xml",
	}

	err := replayRSS(data, feed, d)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "external host")
}
