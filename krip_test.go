package krip

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOnlineUrl(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping online test in short mode")
	}
	var website = "https://www.thepioneerwoman.com/food-cooking/recipes/a11059/restaurant-style-salsa/"
	recipe, err := ScrapeUrl(website, ScrapeOptions{})
	require.NoError(t, err)

	assert.NotEmpty(t, recipe.Url)
	assert.NotEmpty(t, recipe.Name)
	assert.NotEmpty(t, recipe.Images)
	assert.NotEmpty(t, recipe.Ingredients)
	// assert.NotEmpty(t, recipe.Instructions)
	assert.NotEmpty(t, recipe.Publisher)
}

func TestFeedOnline(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping online test in short mode")
	}
	var website = "https://klopotenko.com"
	feed, err := ScrapeFeedUrl(website, FeedOptions{DiscoverySampleSize: 3})
	require.NoError(t, err)

	assert.NotEmpty(t, feed.Url)
	assert.NotEmpty(t, feed.Publisher)
	assert.NotEmpty(t, feed.Entries)
}
