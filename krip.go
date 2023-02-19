package krip

import (
	"errors"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/scraper"
	"github.com/borschtapp/krip/scraper/common"
	"github.com/borschtapp/krip/scraper/website"
	"github.com/borschtapp/krip/utils"
)

var scrapers = map[string]model.Scraper{
	"cookstr":        website.ScrapeCookstr,
	"dinnerly":       website.ScrapeMarleySpoon,
	"fitmencook":     website.ScrapeFitMenCook,
	"kitchenstories": website.ScrapeKitchenStories,
	"marleyspoon":    website.ScrapeMarleySpoon,
}

func RegisterScraper(hostname string, fn model.Scraper) {
	scrapers[hostname] = fn
}

func Scrape(input *model.DataInput) (*model.Recipe, error) {
	recipe := &model.Recipe{}
	if err := common.Scrape(input, recipe); err != nil {
		return nil, err
	}

	alias := utils.HostAlias(input.Url)
	// fill recipe according to the alias scraper implementation
	if aliasScraper, ok := scrapers[alias]; ok {
		if err := aliasScraper(input, recipe); err != nil {
			return nil, errors.New("alias scraper error: " + err.Error())
		}
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
