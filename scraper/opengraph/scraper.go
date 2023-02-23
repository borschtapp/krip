package opengraph

import (
	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/utils"
	"strings"
)

func Scrape(data *model.DataInput, r *model.Recipe) error {
	if data.Document == nil {
		return nil
	}

	head := data.Document.Find("head").First()
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

	if len(r.Images) == 0 {
		if val, ok := head.Find("meta[property='og:image']").Attr("content"); ok {
			r.AddImageUrl(val)
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

	if len(r.Author.Name) == 0 {
		if val, ok := head.Find("head > meta[name='author']").Attr("content"); ok {
			r.Author.Name = utils.CleanupInline(val)
		}
	}

	if len(r.Publisher.Name) == 0 {
		if val, ok := head.Find("meta[property='og:site_name']").Attr("content"); ok && !strings.Contains(val, "http") {
			r.Publisher.Name = utils.CleanupInline(val)
		} else if val := head.Find("title").First().Text(); len(val) != 0 {
			if parts := utils.SplitTitle(val); len(parts) > 1 && len(strings.Fields(parts[len(parts)-1])) < 4 {
				r.Publisher.Name = utils.CleanupInline(parts[len(parts)-1])
			}
		}
	}

	if len(r.Publisher.Url) == 0 {
		r.Publisher.Url = utils.BaseUrl(r.Url)
	}

	return nil
}
