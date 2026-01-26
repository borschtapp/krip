package custom

import (
	"fmt"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/utils"
)

var scrapers = map[string]model.Scraper{
	"dinnerly":        ScrapeMarleySpoon,
	"fitmencook":      ScrapeFitMenCook,
	"gousto":          ScrapeGousto,
	"kitchenstories":  ScrapeKitchenStories,
	"marleyspoon":     ScrapeMarleySpoon,
	"mobile_kptncook": ScrapeKptnCook,
}

func RegisterScraper(hostname string, fn model.Scraper) {
	scrapers[hostname] = fn
}

func Scrape(data *model.DataInput, r *model.Recipe) error {
	alias := utils.HostAlias(data.Url)
	if fn, ok := scrapers[alias]; ok {
		if err := fn(data, r); err != nil {
			return fmt.Errorf("custom scraper error: %w", err)
		}
	}
	return nil
}

var feedScrapers = map[string]model.FeedScraper{
	"hellofresh":  ScrapeHelloFreshFeed,
	"marleyspoon": ScrapeMarleySpoonFeed,
}

func RegisterFeedScraper(hostname string, fn model.FeedScraper) {
	feedScrapers[hostname] = fn
}

func ScrapeFeed(data *model.DataInput, feed *model.Feed) error {
	alias := utils.HostAlias(data.Url)
	if fn, ok := feedScrapers[alias]; ok {
		if err := fn(data, feed); err != nil {
			return fmt.Errorf("custom feed scraper error: %w", err)
		}
		return nil
	}

	return fmt.Errorf("feed scraper not found for %s", alias)
}
