package scraper

import (
	"fmt"

	"github.com/sosodev/duration"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/scraper/custom"
	"github.com/borschtapp/krip/scraper/opengraph"
	"github.com/borschtapp/krip/scraper/rss"
	"github.com/borschtapp/krip/scraper/schema"
	"github.com/borschtapp/krip/utils"
)

func Scrape(data *model.DataInput, r *model.Recipe) error {
	r.Url = data.Url
	if r.Publisher == nil {
		r.Publisher = &model.Organization{}
	}
	if r.Author == nil {
		r.Author = &model.Person{}
	}

	// fill recipe with schema.org/Recipe metadata
	if err := schema.Scrape(data, r); err != nil {
		fmt.Println("schema error: " + err.Error())
	}

	// fill recipe with OpenGraph metadata
	if err := opengraph.Scrape(data, r); err != nil {
		fmt.Println("opengraph error: " + err.Error())
	}

	// fill recipe according to the website scraper implementation
	if err := custom.Scrape(data, r); err != nil {
		fmt.Println("website error: " + err.Error())
	}

	if len(r.Language) == 0 && len(r.Url) != 0 {
		domain := utils.DomainZone(r.Url)
		if lang, ok := utils.LanguageByDomain(domain); ok {
			r.Language = lang
		}
	}

	normalizeRecipe(r)
	return nil
}

func normalizeRecipe(r *model.Recipe) {
	if r.Publisher != nil && len(r.Publisher.Name) == 0 {
		r.Publisher = nil
	}

	if r.Author != nil && (len(r.Author.Name) == 0 || (r.Publisher != nil && r.Author.Name == r.Publisher.Name)) {
		r.Author = nil
	}

	if len(r.CookTime) != 0 && len(r.PrepTime) != 0 && len(r.TotalTime) == 0 {
		r.TotalTime = duration.Format(utils.Parse8601Duration(r.CookTime) + utils.Parse8601Duration(r.PrepTime))
	} else if len(r.TotalTime) != 0 && len(r.CookTime) != 0 && len(r.PrepTime) == 0 {
		prepTime := utils.Parse8601Duration(r.TotalTime) - utils.Parse8601Duration(r.CookTime)
		if prepTime > 0 {
			r.PrepTime = duration.Format(prepTime)
		}
	} else if len(r.TotalTime) != 0 && len(r.PrepTime) != 0 && len(r.CookTime) == 0 {
		cookTime := utils.Parse8601Duration(r.TotalTime) - utils.Parse8601Duration(r.PrepTime)
		if cookTime > 0 {
			r.CookTime = duration.Format(cookTime)
		}
	}
}

func ScrapeFeed(data *model.DataInput, opts ...model.FeedOptions) (*model.Feed, error) {
	var opt model.FeedOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	feed := &model.Feed{
		Url: data.Url,
	}

	if err := scrapeEntries(data, feed); err != nil {
		return nil, err
	}

	if !opt.Quick {
		for _, entry := range feed.Entries {
			if len(entry.Url) == 0 {
				continue
			}

			input, err := UrlInput(entry.Url)
			if err != nil {
				continue
			}
			if err := Scrape(input, entry); err != nil {
				continue
			}
		}
	}

	feed.Entries = filterEntries(feed.Entries, opt)
	return feed, nil
}

func scrapeEntries(data *model.DataInput, feed *model.Feed) error {
	if err := custom.ScrapeFeed(data, feed); err == nil && len(feed.Entries) > 0 {
		return nil
	}

	if err := rss.ScrapeFeed(data, feed); err == nil && len(feed.Entries) > 0 {
		return nil
	}

	if err := schema.ScrapeFeed(data, feed); err == nil && len(feed.Entries) > 0 {
		return nil
	}

	return fmt.Errorf("no entries found")
}

func filterEntries(entries []*model.Recipe, opt model.FeedOptions) []*model.Recipe {
	if opt.MinIngredients == 0 && !opt.RequireImage && !opt.RequireInstructions {
		return entries
	}

	filtered := make([]*model.Recipe, 0, len(entries))
	for _, entry := range entries {
		if opt.MinIngredients > 0 && len(entry.Ingredients) < opt.MinIngredients {
			continue
		}
		if opt.RequireImage && len(entry.Images) == 0 {
			continue
		}
		if opt.RequireInstructions && len(entry.Instructions) == 0 {
			continue
		}
		filtered = append(filtered, entry)
	}
	return filtered
}
