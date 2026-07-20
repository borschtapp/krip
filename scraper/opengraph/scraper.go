package opengraph

import (
	"cmp"
	"strings"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/utils"
)

type documentMeta struct {
	name          string
	description   string
	language      string
	imageURL      string
	publisherName string
	authorName    string
}

func scrapeDocumentMeta(data *model.DataInput) *documentMeta {
	if data.Document == nil {
		return nil
	}

	meta := &documentMeta{}
	head := data.Document.Find("head").First()
	rawTitle := utils.CleanupInline(head.Find("title").First().Text())
	titleParts := utils.SplitTitle(rawTitle)

	if val, ok := utils.FirstMetaContent(head, "og:name", "og:title"); ok {
		meta.name = utils.CleanupInline(val)
	} else if len(titleParts) > 1 {
		meta.name = titleParts[0]
	} else if len(rawTitle) != 0 {
		meta.name = rawTitle
	}

	if val, ok := utils.FirstMetaContent(head, "og:description"); ok {
		meta.description = utils.Cleanup(val)
	}

	if val, ok := utils.FirstMetaContent(head, "og:image"); ok {
		meta.imageURL = val
	}

	if val, ok := utils.FirstMetaContent(head, "og:locale"); ok {
		meta.language = utils.CleanupLang(val)
	} else if val, ok := utils.GetMetaContent(head, "http-equiv", "content-language"); ok {
		meta.language = utils.CleanupLang(val)
	} else if val, ok := data.Document.Find("html").Attr("lang"); ok {
		meta.language = utils.CleanupLang(val)
	}

	if val, ok := utils.FirstMetaContent(head, "og:site_name"); ok && !strings.Contains(val, "http") {
		meta.publisherName = utils.CleanupInline(val)
	} else if len(titleParts) > 1 && len(strings.Fields(titleParts[len(titleParts)-1])) < 4 {
		meta.publisherName = titleParts[len(titleParts)-1]
	}

	if val, ok := utils.FirstMetaContent(head, "author"); ok {
		meta.authorName = utils.CleanupInline(val)
	}

	return meta
}

func Scrape(data *model.DataInput, r *model.Recipe) error {
	meta := scrapeDocumentMeta(data)
	if meta == nil {
		return nil
	}

	r.Name = cmp.Or(r.Name, meta.name)
	r.Description = cmp.Or(r.Description, meta.description)
	r.Language = cmp.Or(r.Language, meta.language)
	if len(r.Images) == 0 && len(meta.imageURL) != 0 {
		r.AddImageUrl(meta.imageURL)
	}

	if r.Author == nil {
		r.Author = &model.Person{}
	}
	if r.Publisher == nil {
		r.Publisher = &model.Organization{}
	}
	r.Author.Name = cmp.Or(r.Author.Name, meta.authorName)
	r.Publisher.Name = cmp.Or(r.Publisher.Name, meta.publisherName)
	r.Publisher.Url = cmp.Or(r.Publisher.Url, utils.BaseUrl(data.Url))
	return nil
}

func ScrapeFeed(data *model.DataInput, feed *model.Feed) error {
	meta := scrapeDocumentMeta(data)
	if meta == nil {
		return nil
	}

	feed.Name = cmp.Or(feed.Name, meta.name)
	feed.Description = cmp.Or(feed.Description, meta.description)
	feed.Language = cmp.Or(feed.Language, meta.language)
	if len(feed.Images) == 0 && len(meta.imageURL) != 0 {
		feed.Images = []*model.ImageObject{{Url: meta.imageURL}}
	}

	if feed.Publisher == nil {
		feed.Publisher = &model.Organization{}
	}
	feed.Publisher.Name = cmp.Or(feed.Publisher.Name, meta.publisherName)
	feed.Publisher.Url = cmp.Or(feed.Publisher.Url, utils.BaseUrl(data.Url))
	return nil
}
