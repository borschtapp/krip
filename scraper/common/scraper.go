package common

import (
	"errors"
	"github.com/sosodev/duration"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/scraper/opengraph"
	"github.com/borschtapp/krip/scraper/schema"
	"github.com/borschtapp/krip/utils"
)

func Scrape(data *model.DataInput, r *model.Recipe) error {
	r.Url = data.Url

	// fill recipe with OpenGraph metadata
	if err := opengraph.Scrape(data, r); err != nil {
		return errors.New("opengraph error: " + err.Error())
	}

	// fill recipe with schema.org/Recipe metadata
	if err := schema.Scrape(data, r); err != nil {
		return errors.New("schema error: " + err.Error())
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
		r.TotalTime = duration.Format(utils.ConvertDuration(r.CookTime) + utils.ConvertDuration(r.PrepTime))
	} else if len(r.TotalTime) != 0 && len(r.CookTime) != 0 && len(r.PrepTime) == 0 {
		prepTime := utils.ConvertDuration(r.TotalTime) - utils.ConvertDuration(r.CookTime)
		if prepTime > 0 {
			r.PrepTime = duration.Format(prepTime)
		}
	} else if len(r.TotalTime) != 0 && len(r.PrepTime) != 0 && len(r.CookTime) == 0 {
		cookTime := utils.ConvertDuration(r.TotalTime) - utils.ConvertDuration(r.PrepTime)
		if cookTime > 0 {
			r.CookTime = duration.Format(cookTime)
		}
	}

	//if len(r.ThumbnailUrl) != 0 && len(r.Images) == 0 {
	//	r.Images = []*model.ImageObject{{Url: r.ThumbnailUrl}}
	//} else if len(r.ThumbnailUrl) == 0 && len(r.Images) != 0 {
	//	r.ThumbnailUrl = r.Images[0].Url
	//}
}
