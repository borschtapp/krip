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

	if len(r.Url) == 0 {
		if val, ok := data.Document.Find("meta[property='og:url']").Attr("content"); ok {
			r.Url = val
		} else if val, ok := data.Document.Find("link[rel='canonical']").Attr("href"); ok {
			r.Url = val
		} else if val, ok := data.Document.Find("link[rel='alternate']").Attr("href"); ok {
			r.Url = val
		}
	}

	if len(r.Name) == 0 {
		if val, ok := data.Document.Find("meta[property='og:name']").Attr("content"); ok {
			r.Name = utils.CleanupInline(val)
		} else if val, ok := data.Document.Find("meta[property='og:title']").Attr("content"); ok {
			r.Name = utils.CleanupInline(val)
		} else if text := data.Document.Find("title").Text(); text != "" {
			r.Name = utils.CleanupInline(val)
		}
	}

	if len(r.Description) == 0 {
		if val, ok := data.Document.Find("meta[property='og:description']").Attr("content"); ok {
			r.Description = utils.Cleanup(val)
		}
	}

	if len(r.ThumbnailUrl) == 0 {
		if val, ok := data.Document.Find("meta[property='og:image']").Attr("content"); ok {
			r.ThumbnailUrl = val
		}
	}

	if len(r.Language) == 0 {
		if val, ok := data.Document.Find("meta[property='og:locale']").Attr("content"); ok {
			r.Language = strings.Split(val, ",")[0]
		} else if attr, ok := data.Document.Find("meta[http-equiv='content-language']").Attr("content"); ok {
			r.Language = strings.Split(attr, ",")[0]
		} else if val, ok = data.Document.Find("html").Attr("lang"); ok {
			r.Language = val
		}
	}

	if r.Publisher == nil {
		if val, ok := data.Document.Find("meta[property='og:site_name']").Attr("content"); ok {
			r.Publisher = &model.Organization{Name: utils.CleanupInline(val), Url: utils.BaseUrl(r.Url)}
		}
	}

	return nil
}
