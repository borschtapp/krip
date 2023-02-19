package website

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/scraper"
	"github.com/borschtapp/krip/scraper/common"
	"github.com/borschtapp/krip/testdata"
)

// this website uses comments and new lines inside a json
func TestKlopotenko(t *testing.T) {
	testdata.OptionallyMockRequests(t)

	input, err := scraper.UrlInput("https://klopotenko.com/salat-z-buryakom-i-vishneu/")
	assert.NoError(t, err)

	recipe := &model.Recipe{}
	assert.NoError(t, common.Scrape(input, recipe))
	testdata.AssertRecipe(t, recipe)
}

func TestKlopotenkoHachapuri(t *testing.T) {
	input, err := scraper.UrlInput("https://klopotenko.com/ru/hachapuri-po-adzharski/")
	assert.NoError(t, err)

	recipe := &model.Recipe{}
	assert.NoError(t, common.Scrape(input, recipe))
	assert.Equal(t, 13, len(recipe.Ingredients))
	assert.Equal(t, 11, len(recipe.Instructions))
}
