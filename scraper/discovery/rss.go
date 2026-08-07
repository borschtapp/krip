package discovery

import (
	"fmt"
	"net/url"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/scraper/rss"
	"github.com/borschtapp/krip/utils"
)

func replayRSS(data *model.DataInput, feed *model.Feed, d *model.DiscoveredFeed) error {
	if d.Selector == "" {
		return fmt.Errorf("rss-link source has no URL")
	}

	baseUrl, err := url.Parse(data.Url)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}
	rssUrl, err := url.Parse(d.Selector)
	if err != nil || rssUrl.Host != baseUrl.Host {
		return fmt.Errorf("rss source points to an external host")
	}

	body, _, err := utils.ExecuteRequest(
		utils.RequestConfig{Method: "GET", URL: d.Selector},
		data.RequestOptions,
	)
	if err != nil {
		return fmt.Errorf("rss fetch failed: %w", err)
	}

	rssInput := &model.DataInput{
		Url:            d.Selector,
		Text:           string(body),
		RequestOptions: data.RequestOptions,
	}

	tmpFeed := &model.Feed{}
	if err := rss.ScrapeFeed(rssInput, tmpFeed); err != nil {
		return fmt.Errorf("rss parse failed: %w", err)
	}

	var entries []*model.Recipe
	for _, e := range tmpFeed.Entries {
		if d.UrlPattern != "" && !utils.UrlMatchesPathPattern(e.Url, d.UrlPattern) {
			continue
		}
		entries = append(entries, e)
	}
	feed.AddEntries(entries)
	return nil
}

func tryRSSLink(data *model.DataInput, feed *model.Feed) (*model.DiscoveredFeed, error) {
	if data.Document == nil {
		return nil, fmt.Errorf("no document")
	}

	sel := `link[rel="alternate"][type="application/rss+xml"], link[rel="alternate"][type="application/atom+xml"]`
	href, exists := data.Document.Find(sel).First().Attr("href")
	if !exists || href == "" {
		return nil, fmt.Errorf("no rss link in head")
	}

	baseUrl, err := url.Parse(data.Url)
	if err != nil {
		return nil, err
	}
	absHref := utils.ToAbsoluteUrl(baseUrl, href)

	// Guard against SSRF: the RSS link href in the page may point to any host.
	rssUrl, err := url.Parse(absHref)
	if err != nil || rssUrl.Host != baseUrl.Host {
		return nil, fmt.Errorf("rss link points to an external host")
	}

	body, _, err := utils.ExecuteRequest(utils.RequestConfig{Method: "GET", URL: absHref}, data.RequestOptions)
	if err != nil {
		return nil, err
	}

	rssInput := &model.DataInput{Url: absHref, Text: string(body), RequestOptions: data.RequestOptions}
	if err := rss.ScrapeFeed(rssInput, feed); err != nil {
		return nil, err
	}
	if len(feed.Entries) == 0 {
		return nil, fmt.Errorf("rss feed empty")
	}

	return &model.DiscoveredFeed{
		Source:          SourceRSSLink,
		Selector:        absHref,
		ConfidenceScore: confidenceFromCount(len(feed.Entries), 1, 20),
	}, nil
}
