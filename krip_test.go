package krip

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOnlineUrl(t *testing.T) {
	t.Skip("Just an example of Url scraping")

	var website = "https://www.thepioneerwoman.com/food-cooking/recipes/a11059/restaurant-style-salsa/"
	recipe, err := ScrapeUrl(website)
	assert.NoError(t, err)

	assert.NotEmpty(t, recipe.Url)
	assert.NotEmpty(t, recipe.Name)
	assert.NotEmpty(t, recipe.Images)
	assert.NotEmpty(t, recipe.Ingredients)
	assert.NotEmpty(t, recipe.Instructions)
	assert.NotEmpty(t, recipe.Publisher)
}
