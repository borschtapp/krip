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

// RequestOptions holds reusable HTTP configuration for all requests in a scrape session.
type RequestOptions struct {
	// Context for cancellation. If nil, context.Background() is used.
	Context context.Context
	// Headers to merge with defaults. Custom values take priority.
	Headers http.Header
	// HttpClient to use. If nil, a default 30s-timeout client is used.
	HttpClient HTTPClient
}

// RecipeFilter holds optional criteria for filtering recipes beyond basic validity.
type RecipeFilter struct {
	// Whether to accept recipes without images.
	OptionalImage bool
	// Whether to accept recipes without a publisher.
	OptionalPublisher bool
	// Whether to accept recipes without ingredients.
	OptionalIngredients bool
	// When set, recipes with fewer than these many ingredients are rejected.
	// Useful to skip prepared/frozen food recipes, which usually have very few ingredients.
	MinIngredients int
	// Whether to accept recipes without instructions.
	OptionalInstructions bool
}

// ScrapeOptions options for scraping a recipe
type ScrapeOptions struct {
	RequestOptions
	RecipeFilter

	// When true, skip scraping the URL from meta tags and rely on the input URL.
	SkipMetaUrl bool
	// When true, skip parsing microdata and rely on other sources. It is a main source of data.
	SkipMicrodata bool
	// When true, skip parsing schemas and rely on other sources.
	SkipSchemaScraper bool
	// When true, skip parsing opengraph and rely on other sources.
	SkipOpenGraphScraper bool
	// When true, skip parsing custom scrapers and rely on other sources.
	SkipCustomScrapers bool
}

// FeedOptions options for feed scraping
type FeedOptions struct {
	ScrapeOptions

	// When true, skip parsing feed meta tags and rely on other sources.
	SkipFeedMeta bool
	// When true, skip parsing RSS feeds and rely on other sources.
	SkipRSSScraper bool
	// When true, only the feed will be scraped without scraping each entry separately. Useful for quick runs.
	SkipEntriesScrape bool
	// When true, skip the universal discovery strategy.
	SkipDiscoveryScraper bool
	// AllowDiscoverySampling fetches 2–3 candidate URLs to confirm they are recipe pages.
	// Disabled by default; enabling adds extra HTTP requests but validates low-confidence results.
	AllowDiscoverySampling bool
	// Discovered, if non-nil, skips discovery and reuses previously discovered container/feed.
	// Populate from a previously returned Feed.Discovered value.
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
