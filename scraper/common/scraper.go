package common

import (
	"errors"

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

	if r.Publisher != nil && len(r.Publisher.Name) == 0 {
		r.Publisher = nil
	}

	if r.Author != nil && (len(r.Author.Name) == 0 || (r.Publisher != nil && r.Author.Name == r.Publisher.Name)) {
		r.Author = nil
	}

	return nil
}
