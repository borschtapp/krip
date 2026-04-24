package scraper

import (
	"errors"
	"log"
	"slices"

	"github.com/sosodev/duration"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/scraper/custom"
	"github.com/borschtapp/krip/scraper/discovery"
	"github.com/borschtapp/krip/scraper/opengraph"
	"github.com/borschtapp/krip/scraper/rss"
	"github.com/borschtapp/krip/scraper/schema"
	"github.com/borschtapp/krip/utils"
)

func Scrape(data *model.DataInput, r *model.Recipe, options model.ScrapeOptions) error {
	r.Url = data.Url
	if r.Publisher == nil {
		r.Publisher = &model.Organization{}
	}
	if r.Author == nil {
		r.Author = &model.Person{}
	}

	if !options.SkipSchemaScraper {
		// fill recipe with schema.org/Recipe metadata
		if err := schema.Scrape(data, r); err != nil {
			log.Println("schema error: " + err.Error())
		}
	}

	if !options.SkipOpenGraphScraper {
		// fill recipe with OpenGraph metadata
		if err := opengraph.Scrape(data, r); err != nil {
			log.Println("OpenGraph error: " + err.Error())
		}
	}

	if !options.SkipCustomScrapers {
		// fill recipe according to the website scraper implementation
		if err := custom.Scrape(data, r); err != nil {
			log.Println("website error: " + err.Error())
		}
	}

	if r.Language == "" && r.Url != "" {
		if lang, ok := utils.LanguageByDomain(utils.DomainZone(r.Url)); ok {
			r.Language = lang
		}
	}

	normalizeRecipe(r)
	return nil
}

func normalizeRecipe(r *model.Recipe) {
	if r.Publisher != nil && r.Publisher.Name == "" {
		r.Publisher = nil
	}

	if r.Author != nil && (r.Author.Name == "" || (r.Publisher != nil && r.Author.Name == r.Publisher.Name)) {
		r.Author = nil
	}

	// clean up the title, remove publisher from it
	if r.Name != "" {
		parts := utils.SplitTitle(r.Name)
		if len(parts) > 1 && r.Publisher != nil && slices.Contains(parts, r.Publisher.Name) {
			parts = slices.DeleteFunc(parts, func(p string) bool {
				return p == r.Publisher.Name
			})

			if len(parts) > 0 {
				r.Name = parts[0]
			}
		}
	}

	switch {
	case r.CookTime != "" && r.PrepTime != "" && r.TotalTime == "":
		r.TotalTime = duration.Format(utils.Parse8601Duration(r.CookTime) + utils.Parse8601Duration(r.PrepTime))
	case r.TotalTime != "" && r.CookTime != "" && r.PrepTime == "":
		if prepTime := utils.Parse8601Duration(r.TotalTime) - utils.Parse8601Duration(r.CookTime); prepTime > 0 {
			r.PrepTime = duration.Format(prepTime)
		}
	case r.TotalTime != "" && r.PrepTime != "" && r.CookTime == "":
		if cookTime := utils.Parse8601Duration(r.TotalTime) - utils.Parse8601Duration(r.PrepTime); cookTime > 0 {
			r.CookTime = duration.Format(cookTime)
		}
	}
}

func ScrapeFeed(data *model.DataInput, feed *model.Feed, options model.FeedOptions) error {
	feed.Url = data.Url
	if feed.Publisher == nil {
		feed.Publisher = &model.Organization{}
	}

	if !options.SkipFeedMeta {
		if err := opengraph.ScrapeFeed(data, feed); err != nil {
			log.Println("OpenGraph error: " + err.Error())
		}
	}

	if err := findEntries(data, feed, options); err != nil {
		return err
	}

	if !options.SkipEntriesScrape {
		for _, entry := range feed.Entries {
			if entry.Url == "" {
				continue
			}

			dataInput, err := UrlInput(entry.Url, options.ScrapeOptions)
			if err != nil {
				continue
			}
			if err := Scrape(dataInput, entry, options.ScrapeOptions); err != nil {
				continue
			}
		}
	}

	initialCount := len(feed.Entries)
	originalEntries := feed.Entries
	feed.Entries = filterEntries(feed.Entries, options)

	if len(feed.Entries) < initialCount {
		discarded := initialCount - len(feed.Entries)
		log.Printf("filtered out %d (out of %d) entries that did not pass validation", discarded, initialCount)

		// If more than 30% were discarded, the source likely mixes recipes with unrelated pages.
		if len(feed.Entries) > 0 && discarded*100/initialCount > 30 {
			if pattern, matched := discovery.RefineByUrlPattern(originalEntries, feed.Entries); pattern != "" {
				log.Printf("url pattern found %q: %d total candidates, %d matches pattern, %d validated before", pattern, initialCount, matched, len(feed.Entries))

				if feed.Discovered != nil && feed.Discovered.UrlPattern == "" {
					feed.Discovered.UrlPattern = pattern
				}
			}
		}
	}

	normalizeFeed(feed)
	return nil
}

func findEntries(data *model.DataInput, feed *model.Feed, options model.FeedOptions) error {
	// Fast path: caller already has a DiscoveredFeed from a previous run.
	if options.Discovered != nil {
		if err := discovery.ReplayDiscovered(data, feed, options.Discovered); err != nil {
			feed.Discovered = nil
			log.Println("discovery replay error: " + err.Error())
		} else if len(feed.Entries) > 0 {
			feed.Discovered = options.Discovered
			return nil
		}
	}

	if !options.SkipCustomScrapers {
		if err := custom.ScrapeFeed(data, feed); err == nil && len(feed.Entries) > 0 {
			return nil
		}
	}

	if !options.SkipRSSScraper {
		if err := rss.ScrapeFeed(data, feed); err == nil && len(feed.Entries) > 0 {
			return nil
		}
	}

	if !options.SkipSchemaScraper {
		if err := schema.ScrapeFeed(data, feed); err == nil && len(feed.Entries) > 0 {
			return nil
		}
	}

	if !options.SkipDiscoveryScraper {
		if err := discovery.ScrapeFeed(data, feed); err == nil && len(feed.Entries) > 0 {
			// Validate low-confidence DOM container results by sampling 2–3 candidate URLs.
			// Returns error early to avoid scraping all links individually.
			if options.AllowDiscoverySampling &&
				feed.Discovered != nil &&
				feed.Discovered.Source == discovery.SourceDOMContainer &&
				feed.Discovered.ConfidenceScore < discovery.AcceptThreshold {
				if !validateDiscoverySample(feed.Entries, options) {
					feed.Entries = nil
					feed.Discovered = nil
					return errors.New("discovery: sampling validation failed (probably there are no recipes on the page)")
				}
			}
			return nil
		}
	}

	return errors.New("no entries found")
}

func filterEntries(entries []*model.Recipe, opt model.FeedOptions) []*model.Recipe {
	filtered := make([]*model.Recipe, 0, len(entries))
	for _, entry := range entries {
		if err := entry.Validate(opt.RecipeFilter); err != nil {
			continue
		}
		filtered = append(filtered, entry)
	}
	return filtered
}

func normalizeFeed(feed *model.Feed) {
	if feed.Publisher != nil && feed.Publisher.Name == "" {
		feed.Publisher = nil
	}
}

// validateDiscoverySample fetches up to 3 candidate URLs (spread across the list)
// and returns true if more than half yield a non-empty recipe.
func validateDiscoverySample(entries []*model.Recipe, options model.FeedOptions) bool {
	n := len(entries)
	if n == 0 {
		return false
	}
	// {0, n/2, n-1} can collapse when n < 3; deduplicate so we don't fetch the same URL twice.
	indices := utils.Deduplicate([]int{0, n / 2, n - 1})

	hits := 0
	tried := 0
	for _, i := range indices {
		if i >= n || entries[i].Url == "" {
			continue
		}

		dataInput, err := UrlInput(entries[i].Url, options.ScrapeOptions)
		if err != nil {
			continue
		}

		r := &model.Recipe{}
		_ = Scrape(dataInput, r, options.ScrapeOptions)
		tried++

		if err := r.Validate(options.RecipeFilter); err == nil {
			hits++
		}
	}

	return tried > 0 && hits*2 > tried
}
