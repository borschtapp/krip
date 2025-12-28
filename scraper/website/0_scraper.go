package website

import (
	"fmt"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/utils"
)

var scrapers = map[string]model.Scraper{
	"archanaskitchen":  ScrapeArchanasKitchen,
	"cookstr":          ScrapeCookstr,
	"dinnerly":         ScrapeMarleySpoon,
	"fitmencook":       ScrapeFitMenCook,
	"gousto":           ScrapeGousto,
	"kitchenstories":   ScrapeKitchenStories,
	"marleyspoon":      ScrapeMarleySpoon,
	"mob":              ScrapeMob,
	"mobile_kptncook":  ScrapeKptnCook,
	"whatsgabycooking": ScrapeWhatsGabyCooking,
}

func RegisterScraper(hostname string, fn model.Scraper) {
	scrapers[hostname] = fn
}

func Scrape(data *model.DataInput, r *model.Recipe) error {
	alias := utils.HostAlias(data.Url)
	if aliasScraper, ok := scrapers[alias]; ok {
		if err := aliasScraper(data, r); err != nil {
			return fmt.Errorf("alias scraper error: %w", err)
		}
	}
	return nil
}
