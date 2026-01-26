package model

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/astappiev/microdata"
	"golang.org/x/net/html"
)

// InputOptions options for pre-processing input
type InputOptions struct {
	SkipUrl    bool
	SkipText   bool
	SkipSchema bool
}

// DataInput represents the input data for the scraper
type DataInput struct {
	Url      string
	Text     string
	RootNode *html.Node           `json:"-"`
	Document *goquery.Document    `json:"-"`
	Schemas  *microdata.Microdata `json:"-"`
}

// FeedOptions options for feed scraping
type FeedOptions struct {
	// When true, only the feed will be scraped, without scraping each entry's url
	Quick bool
	// Filter out recipes with fewer than this number of ingredients (0 = no filter)
	MinIngredients int
	// When true, filter out recipes without an image
	RequireImage bool
	// When true, filter out recipes without instructions
	RequireInstructions bool
}

// Scraper defines a function that fill a recipe from the input data
type Scraper = func(data *DataInput, recipe *Recipe) error

// FeedScraper defines a function that returns a feed from the input data
type FeedScraper = func(data *DataInput, feed *Feed) error
