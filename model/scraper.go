package model

import (
	"context"
	"io"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/astappiev/microdata"
	"golang.org/x/net/html"
)

// HTTPClient is an interface that allows http.Client or safeurl.Client to be used interchangeably.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
	Get(url string) (*http.Response, error)
	Post(url, contentType string, body io.Reader) (*http.Response, error)
}

// RequestOptions holds reusable HTTP configuration for a scrape session.
type RequestOptions struct {
	// Cancellation context; defaults to context.Background().
	Context context.Context
	// Headers merged with defaults; custom values take priority.
	Headers http.Header
	// HTTP client to use; defaults to a 30s-timeout client.
	HttpClient HTTPClient
}

// RecipeFilter holds optional criteria for filtering recipes.
type RecipeFilter struct {
	// Accept recipes without images.
	OptionalImage bool
	// Accept recipes without a publisher.
	OptionalPublisher bool
	// Accept recipes without ingredients.
	OptionalIngredients bool
	// Reject recipes with fewer than this many ingredients; 0 = disabled.
	// Useful to filter out prepared/frozen food recipes.
	MinIngredients int
	// Accept recipes without instructions.
	OptionalInstructions bool
}

// ScrapeOptions options for scraping a recipe
type ScrapeOptions struct {
	RequestOptions
	RecipeFilter

	// Skip scraping the URL from meta tags; uses the input URL instead.
	SkipMetaUrl bool
	// Skip parsing microdata (primary data source, used by schema and OpenGraph scrapers).
	SkipMicrodata bool
	// Skip parsing schemas.
	SkipSchemaScraper bool
	// Skip parsing OpenGraph tags.
	SkipOpenGraphScraper bool
	// Skip custom scrapers.
	SkipCustomScrapers bool
}

// FeedOptions options for feed scraping
type FeedOptions struct {
	ScrapeOptions

	// Skip parsing feed meta tags.
	SkipFeedMeta bool
	// Skip parsing RSS feeds.
	SkipRSSScraper bool
	// Skip the universal discovery strategy.
	SkipDiscoveryScraper bool
	// If > 0, sample up to this many URLs per candidate group during DOM discovery.
	// Failed groups are skipped; confirmed entries are not re-fetched later.
	DiscoverySampleSize int
	// Skip scraping individual entries; useful for quick runs.
	SkipEntriesScrape bool
	// Max entries to scrape individually; 0 = default (20), negative = no cap.
	MaxEntriesForScrape int
	// If non-nil, skip discovery and reuse a previously discovered container/feed.
	// Populate from a prior Feed.Discovered value.
	Discovered *DiscoveredFeed
}

// DataInput represents the input data for the scraper
type DataInput struct {
	Url            string
	Text           string
	RootNode       *html.Node           `json:"-"`
	Document       *goquery.Document    `json:"-"`
	Schemas        *microdata.Microdata `json:"-"`
	RequestOptions RequestOptions       `json:"-"`
}

// Scraper defines a function that fill a recipe from the input data
type Scraper = func(data *DataInput, recipe *Recipe) error

// FeedScraper defines a function that returns a feed from the input data
type FeedScraper = func(data *DataInput, feed *Feed) error
