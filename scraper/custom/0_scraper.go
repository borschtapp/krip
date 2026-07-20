package custom

import (
	"fmt"
	"log"
	"runtime/debug"

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

func recoverPanic(label, alias string, errp *error) {
	if p := recover(); p != nil {
		log.Printf("%s panic for %s: %v\n%s", label, alias, p, debug.Stack())
		*errp = fmt.Errorf("%s panic for %s: %v", label, alias, p)
	}
}

func Scrape(data *model.DataInput, r *model.Recipe) (err error) {
	alias := utils.HostAlias(data.Url)
	fn, ok := scrapers[alias]
	if !ok {
		return nil
	}
	defer recoverPanic("custom scraper", alias, &err)
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
	defer recoverPanic("custom feed scraper", alias, &err)
	if err = fn(data, feed); err != nil {
		return fmt.Errorf("custom feed scraper error: %w", err)
	}
	return nil
}
