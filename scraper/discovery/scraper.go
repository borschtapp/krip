package discovery

import (
	"fmt"
	"net/url"

	"github.com/borschtapp/krip/model"
)

// Source constants for DiscoveredFeed.Source.
const (
	SourceRSSLink      model.DiscoverySource = "rss-link"
	SourceSitemap      model.DiscoverySource = "sitemap"
	SourceDOMContainer model.DiscoverySource = "dom-container"
)

// ScrapeFeed runs the universal discovery pipeline against the given DataInput.
// On success, feed.Entries will contain stub recipes (Url only) and feed.Discovered will describe how they were found.
func ScrapeFeed(data *model.DataInput, feed *model.Feed) error {
	// First, try if there is an RSS/Atom <link rel="alternate"> in the page <head>.
	if d, err := tryRSSLink(data, feed); err == nil && d != nil {
		feed.Discovered = d
		return nil
	}

	baseUrl, err := url.Parse(data.Url)
	if err != nil {
		return fmt.Errorf("discovery: invalid base URL: %w", err)
	}

	// Stage 1: DOM container scoring (zero extra requests)
	if d, err := tryDOMScoring(data, feed, baseUrl); err == nil && d != nil {
		feed.Discovered = d
		return nil
	}

	// Stage 2: Sitemap (1–3 extra requests)
	if d, err := trySitemap(data, feed, baseUrl, data.RequestOptions); err == nil && d != nil {
		feed.Discovered = d
		return nil
	}

	return fmt.Errorf("discovery: no entries found")
}

// ReplayDiscovered replays a previously discovered feed configuration.
func ReplayDiscovered(data *model.DataInput, feed *model.Feed, d *model.DiscoveredFeed) error {
	switch d.Source {
	case SourceRSSLink:
		return replayRSS(data, feed, d)

	case SourceSitemap:
		return replaySitemap(data, feed, d)

	case SourceDOMContainer:
		return replayDOMScoring(data, feed, d)

	default:
		feed.Discovered = nil
		return fmt.Errorf("unknown discovery source: %s", d.Source)
	}
}
