package custom_test

import (
	"testing"

	"github.com/borschtapp/krip"
	"github.com/stretchr/testify/assert"
)

func TestKitchenStoriesOnline(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping online test in short mode")
	}

	var website = "https://www.kitchenstories.com/de/rezepte/pochierter-kabeljau-in-tomatensosse"
	recipe, err := krip.ScrapeUrl(website)
	assert.NoError(t, err)
	assert.True(t, recipe.IsValid())
	t.Log(recipe.String())
}
