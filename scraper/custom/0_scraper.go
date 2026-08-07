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
	"hellofresh":      ScrapeHelloFresh,
	"kitchenstories":  ScrapeKitchenStories,
	"marleyspoon":     ScrapeMarleySpoon,
	"mobile_kptncook": ScrapeKptnCook,
}

func RegisterScraper(hostname string, fn model.Scraper) {
	scrapers[hostname] = fn
}

func Scrape(data *model.DataInput, r *model.Recipe) (err error) {
	alias := utils.HostAlias(data.Url)
	fn, ok := scrapers[alias]
	if !ok {
		return nil
	}
	defer utils.RecoverPanic(fmt.Sprintf("custom scraper for %s", alias), &err)
	if err = fn(data, r); err != nil {
		return fmt.Errorf("custom scraper error: %w", err)
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

func ScrapeFeed(data *model.DataInput, feed *model.Feed) (err error) {
	alias := utils.HostAlias(data.Url)
	fn, ok := feedScrapers[alias]
	if !ok {
		return nil
	}
	defer utils.RecoverPanic(fmt.Sprintf("custom feed scraper for %s", alias), &err)
	if err = fn(data, feed); err != nil {
		return fmt.Errorf("custom feed scraper error: %w", err)
	}
	return nil
}
