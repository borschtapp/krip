package krip

import (
	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/scraper"
	"github.com/borschtapp/krip/scraper/custom"
)

func RegisterScraper(hostname string, fn model.Scraper) {
	custom.RegisterScraper(hostname, fn)
}

func RegisterFeedScraper(hostname string, fn model.FeedScraper) {
	custom.RegisterFeedScraper(hostname, fn)
}

// Scrape scrapes a recipe from the input
func Scrape(input *model.DataInput) (*model.Recipe, error) {
	recipe := &model.Recipe{}
	if err := scraper.Scrape(input, recipe); err != nil {
		return nil, err
	}
	return recipe, nil
}

// ScrapeFile reads content and scrapes a recipe from the file
func ScrapeFile(fileName string, options model.ScrapeOptions) (*model.Recipe, error) {
	input, err := scraper.FileInput(fileName, options)
	if err != nil {
		return nil, err
	}

	return Scrape(input)
}

// ScrapeUrl retrieves and scrapes a recipe from the url
func ScrapeUrl(url string, options model.ScrapeOptions) (*model.Recipe, error) {
	input, err := scraper.UrlInput(url, options)
	if err != nil {
		return nil, err
	}

	return Scrape(input)
}

// ScrapeFeed scrapes a feed of recipes from the input
func ScrapeFeed(input *model.DataInput, options model.FeedOptions) (*model.Feed, error) {
	return scraper.ScrapeFeed(input, options)
}

// ScrapeFeedUrl retrieves and scrapes a feed of recipes from the url
func ScrapeFeedUrl(url string, options model.FeedOptions) (*model.Feed, error) {
	input, err := scraper.UrlInput(url, options.ScrapeOptions)
	if err != nil {
		return nil, err
	}

	return ScrapeFeed(input, options)
}

// Exported types below

type Person = model.Person
type Organization = model.Organization
type PropertyValue = model.PropertyValue
type HowToTool = model.HowToTool
type HowToStep = model.HowToStep
type HowToSection = model.HowToSection
type NutritionInformation = model.NutritionInformation
type AggregateRating = model.AggregateRating
type ImageObject = model.ImageObject
type VideoObject = model.VideoObject
type Recipe = model.Recipe
type Feed = model.Feed

type RequestOptions = model.RequestOptions
type ScrapeOptions = model.ScrapeOptions
type FeedOptions = model.FeedOptions
type DataInput = model.DataInput
type Scraper = model.Scraper
type FeedScraper = model.FeedScraper
