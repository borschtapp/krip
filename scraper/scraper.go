package scraper

import (
	"errors"
	"log"
	"slices"
	"strings"

	"github.com/sosodev/duration"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/scraper/custom"
	"github.com/borschtapp/krip/scraper/discovery"
	"github.com/borschtapp/krip/scraper/opengraph"
	"github.com/borschtapp/krip/scraper/rss"
	"github.com/borschtapp/krip/scraper/schema"
	"github.com/borschtapp/krip/utils"
)

// maxEntriesForScrape is the default cap for individual entry scraping per feed.
// Entries beyond this index are kept as stub entries (URL only).
const maxEntriesForScrape = 20

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
			log.Printf("schema error: %v", err)
		}
	}

	if !options.SkipOpenGraphScraper {
		// fill recipe with OpenGraph metadata
		if err := opengraph.Scrape(data, r); err != nil {
			log.Printf("OpenGraph error: %v", err)
		}
	}

	if !options.SkipCustomScrapers {
		// fill recipe according to the website scraper implementation
		if err := custom.Scrape(data, r); err != nil {
			log.Printf("website error: %v", err)
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

	// Remove publisher from the title.
	if r.Name != "" {
		parts := utils.SplitTitle(r.Name)
		if len(parts) > 1 && r.Publisher != nil && slices.Contains(parts, r.Publisher.Name) {
			parts = slices.DeleteFunc(parts, func(p string) bool {
				return strings.EqualFold(p, r.Publisher.Name)
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
			log.Printf("feed: OpenGraph error: %v", err)
		}
	}

	scrapedURLs := make(map[string]bool)
	if err := findEntries(data, feed, options, scrapedURLs); err != nil {
		return err
	}

	log.Printf("feed: %d entries found from %s", len(feed.Entries), feed.Url)
	limit := options.MaxEntriesForScrape
	switch {
	case limit == 0:
		limit = maxEntriesForScrape
	case limit < 0:
		limit = len(feed.Entries)
	}

	if !options.SkipEntriesScrape {
		log.Println("feed: scraping individual entries for more complete data")
		toScrape := feed.Entries
		if len(feed.Entries) > limit {
			toScrape = feed.Entries[:limit]
			log.Printf("feed: scraping top %d of %d entries; the rest remain as stubs", limit, len(feed.Entries))
		}

		for _, entry := range toScrape {
			if options.ContextDone() {
				break
			}
			if entry.Url == "" || scrapedURLs[entry.Url] {
				continue
			}

			scrapedURLs[entry.Url] = true
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
	feed.Entries = filterEntries(feed.Entries, options, scrapedURLs)

	if len(feed.Entries) < initialCount {
		discarded := initialCount - len(feed.Entries)
		log.Printf("feed: filtered out %d (out of %d) entries that did not pass validation", discarded, initialCount)

		// If many entries were discarded, try to find a URL pattern.
		if len(feed.Entries) > 0 && discarded*100/initialCount > 30 {
			if pattern, matched := discovery.RefineByUrlPattern(originalEntries, feed.Entries); pattern != "" {
				log.Printf("feed: url pattern found %q: %d total candidates, %d matches pattern, %d validated before", pattern, initialCount, matched, len(feed.Entries))

				if feed.Discovered != nil && feed.Discovered.UrlPattern == "" {
					feed.Discovered.UrlPattern = pattern
				}
			}
		}
	}

	normalizeFeed(feed)
	return nil
}

func findEntries(data *model.DataInput, feed *model.Feed, options model.FeedOptions, scrapedURLs map[string]bool) error {
	// Fast path: caller already has a DiscoveredFeed from a previous run.
	if options.Discovered != nil {
		if err := discovery.ReplayDiscovered(data, feed, options.Discovered); err != nil {
			feed.Discovered = nil
			log.Printf("feed: discovery replay error: %v", err)
		} else if len(feed.Entries) > 0 {
			feed.Discovered = options.Discovered
			return nil
		}
	}

	if !options.SkipCustomScrapers {
		if err := custom.ScrapeFeed(data, feed); err != nil {
			log.Printf("custom feed scraper error: %v", err)
		} else if len(feed.Entries) > 0 {
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
			if len(feed.Entries) == 1 && !feed.Entries[0].IsValid() {
				// Single invalid entry is likely a category page
				feed.Entries = nil
			} else {
				return nil
			}
		}
	}

	// Try universal discovery: DOM scoring, RSS link, then sitemap.
	if !options.SkipDiscoveryScraper {
		// prescraped holds confirmed recipes from group validation sampling.
		// It is populated inside the closure and read after ScrapeFeed returns,
		// so we only mark URLs that ended up in the committed group.
		var prescraped map[string]*model.Recipe
		sampling := discovery.SamplingOptions{}
		if options.DiscoverySampleSize > 0 {
			prescraped = make(map[string]*model.Recipe)
			sampling = discovery.SamplingOptions{
				SampleSize: options.DiscoverySampleSize,
				// Fetches URLs one at a time and stops as soon as the accept/reject
				// verdict used by discovery's selectBestGroup (hits*2 > n) is
				// mathematically decided, so the remaining sample doesn't need to be
				// fetched. n stays the full requested sample size regardless of how
				// many were actually tried, matching applyGroupValidation's sampleCount.
				Validator: func(urls []string) []*model.Recipe {
					n := len(urls)
					var confirmed []*model.Recipe
					tried := 0
					for _, u := range urls {
						if options.ContextDone() {
							break
						}
						tried++
						dataInput, err := UrlInput(u, options.ScrapeOptions)
						if err == nil {
							r := &model.Recipe{Url: u}
							_ = Scrape(dataInput, r, options.ScrapeOptions)
							if r.Validate(options.RecipeFilter) == nil {
								prescraped[u] = r
								confirmed = append(confirmed, r)
							}
						}

						hits := len(confirmed)
						if hits*2 > n || (tried-hits) >= (n+1)/2 {
							break
						}
					}
					return confirmed
				},
			}
		}

		scoring := discovery.ScoringOptions{
			AcceptThreshold: options.DOMAcceptThreshold,
			SampleThreshold: options.DOMSampleThreshold,
			MinGroupSize:    options.DOMMinGroupSize,
			MaxGroupsCheck:  options.DOMMaxGroupsCheck,
		}
		if err := discovery.ScrapeFeed(data, feed, sampling, scoring); err == nil && len(feed.Entries) > 0 {
			if prescraped != nil {
				for _, e := range feed.Entries {
					if _, ok := prescraped[e.Url]; ok {
						scrapedURLs[e.Url] = true
					}
				}
			}
			return nil
		}
	}

	return errors.New("no entries found")
}

func filterEntries(entries []*model.Recipe, opt model.FeedOptions, scrapedURLs map[string]bool) []*model.Recipe {
	filtered := make([]*model.Recipe, 0, len(entries))
	for _, entry := range entries {
		if scrapedURLs != nil && entry.Url != "" && !scrapedURLs[entry.Url] {
			filtered = append(filtered, entry)
			continue
		}
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
