package website

import (
	"errors"
	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/utils"
)

var scrapers = map[string]model.Scraper{
	"cookstr":          ScrapeCookstr,
	"dinnerly":         ScrapeMarleySpoon,
	"fitmencook":       ScrapeFitMenCook,
	"kitchenstories":   ScrapeKitchenStories,
	"marleyspoon":      ScrapeMarleySpoon,
	"archanaskitchen":  ScrapeArchanasKitchen,
	"whatsgabycooking": ScrapeWhatsGabyCooking,
	"mob":              ScrapeMob,
}

func RegisterScraper(hostname string, fn model.Scraper) {
	scrapers[hostname] = fn
}

func Scrape(data *model.DataInput, r *model.Recipe) error {
	alias := utils.HostAlias(data.Url)
	if aliasScraper, ok := scrapers[alias]; ok {
		if err := aliasScraper(data, r); err != nil {
			return errors.New("alias scraper error: " + err.Error())
		}
	}
	return nil
}
