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

// ScrapeOptions options for scraping a recipe
type ScrapeOptions struct {
	RequestOptions

	// When true, skip scraping the URL from meta tags and rely on the input URL.
	SkipMetaUrl bool
	// When true, skip parsing microdata and rely on other sources. It is a main source of data.
	SkipMicrodata bool
}

// FeedOptions options for feed scraping
type FeedOptions struct {
	ScrapeOptions

	// When true, only the feed will be scraped, without scraping each entry's url
	Quick bool
	// Filter out recipes with fewer than this number of ingredients (0 = no filter)
	MinIngredients int
	// When true, filter out recipes without an image
	RequireImage bool
	// When true, filter out recipes without instructions
	RequireInstructions bool
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
