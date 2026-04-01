package discovery

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestContainsRecipeKeyword(t *testing.T) {
	assert.True(t, containsRecipeKeyword("https://example.com/recipe-sitemap.xml"))
	assert.True(t, containsRecipeKeyword("https://example.com/food-sitemap.xml"))
	assert.False(t, containsRecipeKeyword("https://example.com/page-sitemap.xml"))
	assert.False(t, containsRecipeKeyword("https://example.com/sitemap.xml"))
}
