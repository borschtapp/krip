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

// Scraper defines a function that fill a recipe from the input data
type Scraper = func(data *DataInput, r *Recipe) error
