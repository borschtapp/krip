package website

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/scraper"
	"github.com/borschtapp/krip/scraper/common"
	"github.com/borschtapp/krip/test"
)

func TestFitMenCook(t *testing.T) {
	test.OptionallyMockRequests(t)

	input, err := scraper.UrlInput("https://fitmencook.com/healthy-chili-recipe/")
	assert.NoError(t, err)

	recipe := &model.Recipe{}
	assert.NoError(t, common.Scrape(input, recipe))
	assert.NoError(t, ScrapeFitMenCook(input, recipe))
	test.AssertRecipe(t, recipe)
}
