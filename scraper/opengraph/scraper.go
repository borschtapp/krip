package opengraph

import (
	"strings"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/utils"
)

func Scrape(data *model.DataInput, r *model.Recipe) error {
	if data.Document == nil {
		return nil
	}

	head := data.Document.Find("head").First()
	if len(r.Url) == 0 {
		if val, ok := head.Find("meta[property='og:url']").Attr("content"); ok && utils.IsAbsolute(val) {
			r.Url = val
		} else if val, ok := head.Find("link[rel='canonical']").Attr("href"); ok && utils.IsAbsolute(val) {
			r.Url = val
		} else if val, ok := head.Find("link[rel='alternate']").Attr("href"); ok && utils.IsAbsolute(val) {
			r.Url = val
		} else {
			r.Url = data.Url
		}
	}

	if len(r.Name) == 0 {
		if val, ok := head.Find("meta[property='og:name']").Attr("content"); ok {
			r.Name = utils.CleanupInline(val)
		} else if val, ok := head.Find("meta[property='og:title']").Attr("content"); ok {
			r.Name = utils.CleanupInline(val)
		} else if val := head.Find("title").First().Text(); len(val) != 0 {
			r.Name = utils.CleanupInline(val)
		}
	}

	if len(r.Description) == 0 {
		if val, ok := head.Find("meta[property='og:description']").Attr("content"); ok {
			r.Description = utils.Cleanup(val)
		}
	}

	if len(r.ThumbnailUrl) == 0 {
		if val, ok := head.Find("meta[property='og:image']").Attr("content"); ok {
			r.ThumbnailUrl = val
		}
	}

	if len(r.Language) == 0 {
		if val, ok := head.Find("meta[property='og:locale']").Attr("content"); ok {
			r.Language = utils.CleanupLang(val)
		} else if val, ok := head.Find("meta[http-equiv='content-language']").Attr("content"); ok {
			r.Language = utils.CleanupLang(val)
		} else if val, ok = data.Document.Find("html").Attr("lang"); ok {
			r.Language = utils.CleanupLang(val)
		}
	}

	if r.Author == nil {
		if val, ok := head.Find("head > meta[name='author']").Attr("content"); ok {
			r.Author = &model.Person{Name: utils.CleanupInline(val)}
		}
	}

	if r.Publisher == nil {
		r.Publisher = &model.Organization{}

		if val, ok := head.Find("meta[property='og:site_name']").Attr("content"); ok && !strings.Contains(val, "http") {
			r.Publisher.Name = utils.CleanupInline(val)
		} else if val := head.Find("title").First().Text(); len(val) != 0 {
			if parts := utils.SplitTitle(val); len(parts) > 1 && len(strings.Fields(parts[len(parts)-1])) < 4 {
				r.Publisher.Name = utils.CleanupInline(parts[len(parts)-1])
			}
		}

		if len(r.Publisher.Url) == 0 {
			r.Publisher.Url = utils.BaseUrl(r.Url)
		}
	}

	return nil
}
