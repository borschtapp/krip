package custom_test

import (
	"testing"

	"github.com/borschtapp/krip"
	"github.com/stretchr/testify/assert"
)

func TestHelloFreshOnline(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping online test in short mode")
	}

	var website = "https://www.hellofresh.de/recipes/gnocchi-salat-mit-selbst-gemachtem-caesar-dressing-65144891aab557d393d8c000"
	recipe, err := krip.ScrapeUrl(website)
	assert.NoError(t, err)
	assert.True(t, recipe.IsValid())
	t.Log(recipe.String())
}

func TestHelloFreshFeedOnline(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping online test in short mode")
	}

	var website = "https://www.hellofresh.de/menus"
	feed, err := krip.ScrapeFeedUrl(website)
	assert.NoError(t, err)
	assert.NotNil(t, feed)

	assert.NotEmpty(t, feed.Url)
	assert.NotEmpty(t, feed.Entries)
	for _, entry := range feed.Entries {
		assert.False(t, entry.IsEmpty())
	}
	t.Log(feed.String())
}
