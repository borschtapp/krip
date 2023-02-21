package krip

import (
	"github.com/borschtapp/krip/testdata"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOnlineUrl(t *testing.T) {
	t.Skip("Just an example")

	var website = "https://www.thepioneerwoman.com/food-cooking/recipes/a11059/restaurant-style-salsa/"
	recipe, err := ScrapeUrl(website)
	assert.NoError(t, err)

	assert.NotEmpty(t, recipe.Name)
	testdata.AssertRecipe(t, recipe)
}

func TestHtmlFile(t *testing.T) {
	t.Skip("Just an example")

	var website = "kitchenstories"
	recipe, err := ScrapeFile(testdata.WebsitesDir + website + testdata.HtmlExt)
	assert.NoError(t, err)

	testdata.AssertJson(t, recipe, testdata.RecipesDir+website)
}
