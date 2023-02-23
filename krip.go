package krip

import (
	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/scraper"
	"github.com/borschtapp/krip/scraper/common"
	"github.com/borschtapp/krip/scraper/website"
)

func RegisterScraper(hostname string, fn model.Scraper) {
	website.RegisterScraper(hostname, fn)
}

func Scrape(input *model.DataInput) (*model.Recipe, error) {
	recipe := &model.Recipe{}
	if err := common.Scrape(input, recipe); err != nil {
		return nil, err
	}
	return recipe, nil
}

// ScrapeFile reads content and scrapes a recipe from the file
func ScrapeFile(fileName string) (*model.Recipe, error) {
	input, err := scraper.FileInput(fileName, model.InputOptions{})
	if err != nil {
		return nil, err
	}

	return Scrape(input)
}

// ScrapeUrl retrieves and scrapes a recipe from the url
func ScrapeUrl(url string) (*model.Recipe, error) {
	input, err := scraper.UrlInput(url)
	if err != nil {
		return nil, err
	}

	return Scrape(input)
}
