package model

import (
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"

	"github.com/astappiev/microdata"
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
	RootNode *html.Node        `json:"-"`
	Document *goquery.Document `json:"-"`
	Schema   *microdata.Item   `json:"-"`
}

// Scraper defines a function that fill a recipe from the input data
type Scraper = func(data *DataInput, r *Recipe) error

// Krip represents an accumulated scraper data
type Krip struct {
	Input       *DataInput    `json:"input,omitempty"`
	Recipe      *Recipe       `json:"recipe,omitempty"`
	Ingredients []*Ingredient `json:"ingredients,omitempty"`
}
