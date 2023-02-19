package website

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/scraper"
	"github.com/borschtapp/krip/scraper/common"
	"github.com/borschtapp/krip/test"
)

func TestMarleySpoon(t *testing.T) {
	test.OptionallyMockRequests(t)

	input, err := scraper.UrlInput("https://marleyspoon.de/menu/113813-glasierte-veggie-burger-mit-roestkartoffeln-und-apfel-gurken-salat")
	assert.NoError(t, err)

	recipe := &model.Recipe{}
	assert.NoError(t, common.Scrape(input, recipe))
	assert.NoError(t, ScrapeMarleySpoon(input, recipe))
	test.AssertRecipe(t, recipe)
}

func TestDinnerly(t *testing.T) {
	test.OptionallyMockRequests(t)

	input, err := scraper.UrlInput("https://dinnerly.de/menu/114391-koestliches-haehnchen-mit-erdnusssauce-in-einer-brokkoli-nudelpfanne")
	assert.NoError(t, err)

	recipe := &model.Recipe{}
	assert.NoError(t, common.Scrape(input, recipe))
	assert.NoError(t, ScrapeMarleySpoon(input, recipe))
	test.AssertRecipe(t, recipe)
}
