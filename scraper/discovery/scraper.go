package discovery

import (
	"fmt"
	"log"
	"net/url"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/utils"
)

// Source constants for DiscoveredFeed.Source.
const (
	SourceRSSLink      model.DiscoverySource = "rss-link"
	SourceSitemap      model.DiscoverySource = "sitemap"
	SourceDOMContainer model.DiscoverySource = "dom-container"
)

// GroupValidator samples a slice of candidate URLs and returns the subset that are
// confirmed recipe pages (already scraped, ready to merge into feed entries).
type GroupValidator func(urls []string) []*model.Recipe

// SamplingOptions configures optional URL validation during DOM discovery.
// The zero value disables sampling entirely.
type SamplingOptions struct {
	// Validator is called with a sample of candidate URLs; nil disables validation.
	Validator GroupValidator
	// SampleSize is the number of URLs to sample per candidate group; 0 disables sampling.
	SampleSize int
}

// ScoringOptions tunes the DOM heuristic scoring thresholds.
// Zero values fall back to the package-level defaults.
type ScoringOptions struct {
	// AcceptThreshold overrides the confidence score above which a group is accepted without sampling.
	AcceptThreshold float64
	// SampleThreshold overrides the confidence score above which sampling is required.
	SampleThreshold float64
	// MinGroupSize overrides the minimum number of links a finalized container group must have to be considered.
	MinGroupSize int
	// MinSiblingsToMerge overrides the minimum number of sibling child groups required before mergeSiblingGroups collapses them into one container.
	MinSiblingsToMerge int
	// MaxGroupsCheck overrides the maximum number of candidate groups to validate via sampling.
	MaxGroupsCheck int
}

// ScrapeFeed runs the discovery pipeline against the given DataInput.
// On success, feed.Entries will contain stub recipes (Url only) and feed.Discovered will describe how they were found.
// When sampling.Validator is set, it is called with a sample of URLs from each DOM candidate group before committing it.
// Groups where fewer than half the sampled URLs are recipes are skipped; the next-best group is tried instead.
func ScrapeFeed(data *model.DataInput, feed *model.Feed, sampling SamplingOptions, scoring ScoringOptions) error {
	baseUrl, err := url.Parse(data.Url)
	if err != nil {
		return fmt.Errorf("discovery: invalid base URL: %w", err)
	}

	// Stage 0: DOM container scoring (no extra requests, optionally for validation)
	log.Println("discovery: trying DOM container scoring")
	if d, err := runStage(feed, "DOM container scoring", func() (*model.DiscoveredFeed, error) {
		return tryDOMScoring(data, feed, baseUrl, sampling, scoring)
	}); err == nil && d != nil {
		feed.Discovered = d
		log.Printf("discovery: DOM container found with confidence score %.4f", d.ConfidenceScore)
		return nil
	} else if err != nil {
		log.Printf("discovery: DOM container scoring error: %v", err)
	}

	// Stage 1: RSS/Atom <link rel="alternate"> feed (from the page <head>, 1 extra request)
	log.Println("discovery: trying RSS/Atom link in page head")
	if d, err := runStage(feed, "RSS/Atom link discovery", func() (*model.DiscoveredFeed, error) {
		return tryRSSLink(data, feed)
	}); err == nil && d != nil {
		feed.Discovered = d
		log.Println("discovery: RSS/Atom link found")
		return nil
	} else if err != nil {
		log.Printf("discovery: RSS/Atom link error: %v", err)
	}

	// Stage 2: Sitemap (1–3 extra requests)
	log.Println("discovery: trying sitemap")
	if d, err := runStage(feed, "sitemap discovery", func() (*model.DiscoveredFeed, error) {
		return trySitemap(data, feed, baseUrl, data.RequestOptions)
	}); err == nil && d != nil {
		feed.Discovered = d
		log.Println("discovery: sitemap found")
		return nil
	} else if err != nil {
		log.Printf("discovery: sitemap error: %v", err)
	}

	return fmt.Errorf("discovery: no entries found")
}

// runStage invokes fn, recovering any panic into an error and rolling back any
// feed.Entries it may have added before panicking. This keeps the invariant
// that ordinary stage failures already provide — a failed stage leaves no
// partial state for the next stage to inherit — true for panics as well,
// since a panic unwinds the stack without undoing prior mutations to feed.
func runStage(feed *model.Feed, label string, fn func() (*model.DiscoveredFeed, error)) (d *model.DiscoveredFeed, err error) {
	entriesBefore := len(feed.Entries)
	defer func() {
		if err != nil {
			feed.Entries = feed.Entries[:entriesBefore]
		}
	}()
	defer utils.RecoverPanic(label, &err)
	return fn()
}

// ReplayDiscovered replays a previously discovered feed configuration.
func ReplayDiscovered(data *model.DataInput, feed *model.Feed, d *model.DiscoveredFeed) error {
	_, err := runStage(feed, fmt.Sprintf("discovery replay (%s)", d.Source), func() (*model.DiscoveredFeed, error) {
		switch d.Source {
		case SourceRSSLink:
			return nil, replayRSS(data, feed, d)

		case SourceSitemap:
			return nil, replaySitemap(data, feed, d)

		case SourceDOMContainer:
			return nil, replayDOMScoring(data, feed, d)

		default:
			feed.Discovered = nil
			return nil, fmt.Errorf("unknown discovery source: %s", d.Source)
		}
	})
	return err
}
