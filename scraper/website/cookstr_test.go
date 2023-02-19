package website

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/scraper"
	"github.com/borschtapp/krip/scraper/common"
	"github.com/borschtapp/krip/testdata"
)

func TestCookstr(t *testing.T) {
	testdata.OptionallyMockRequests(t)

	input, err := scraper.UrlInput("https://www.cookstr.com/recipes/thai-style-pot-roast-with-fat-noodles")
	assert.NoError(t, err)

	recipe := &model.Recipe{}
	assert.NoError(t, common.Scrape(input, recipe))
	assert.NoError(t, ScrapeCookstr(input, recipe))
	testdata.AssertRecipe(t, recipe)
}
