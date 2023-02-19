package model

import (
	"encoding/json"
	"time"
)

// Person according to https://schema.org/Person
type Person struct {
	Name        string `json:"name,omitempty"`
	KnowsAbout  string `json:"knowsAbout,omitempty"`
	JobTitle    string `json:"jobTitle,omitempty"`
	Description string `json:"description,omitempty"`
	Url         string `json:"url,omitempty"`
	Image       string `json:"image,omitempty"`
}

// Organization according to https://schema.org/Organization
type Organization struct {
	Name string `json:"name,omitempty"`
	Url  string `json:"url,omitempty"`
	Logo string `json:"logo,omitempty"`
}

// HowToStep a step in the instructions https://schema.org/HowToStep
type HowToStep struct {
	Name  string `json:"name,omitempty"`
	Text  string `json:"text,omitempty"`
	Url   string `json:"url,omitempty"`
	Image string `json:"image,omitempty"`
	Video string `json:"video,omitempty"`
}

// HowToSection a subgrouping of steps in the instructions https://schema.org/HowToSection
type HowToSection struct {
	HowToStep
	Steps []*HowToStep `json:"itemListElement,omitempty"`
}

// NutritionInformation according to https://schema.org/NutritionInformation
type NutritionInformation struct {
	Calories              string `json:"calories,omitempty"`              // The number of calories.
	ServingSize           string `json:"servingSize,omitempty"`           // The serving size, in terms of the number of volume or mass.
	CarbohydrateContent   string `json:"carbohydrateContent,omitempty"`   // The number of grams of carbohydrates.
	CholesterolContent    string `json:"cholesterolContent,omitempty"`    // The number of milligrams of cholesterol.
	FatContent            string `json:"fatContent,omitempty"`            // The number of grams of fat.
	FiberContent          string `json:"fiberContent,omitempty"`          // The number of grams of fiber.
	ProteinContent        string `json:"proteinContent,omitempty"`        // The number of grams of protein.
	SaturatedFatContent   string `json:"saturatedFatContent,omitempty"`   // The number of grams of saturated fat.
	SodiumContent         string `json:"sodiumContent,omitempty"`         // The number of milligrams of sodium.
	SugarContent          string `json:"sugarContent,omitempty"`          // The number of grams of sugar.
	TransFatContent       string `json:"transFatContent,omitempty"`       // The number of grams of trans fat.
	UnsaturatedFatContent string `json:"unsaturatedFatContent,omitempty"` // The number of grams of unsaturated fat.
}

// AggregateRating represents the average rating based on multiple ratings or reviews https://schema.org/AggregateRating
type AggregateRating struct {
	ReviewCount int     `json:"reviewCount,omitempty"`
	RatingCount int     `json:"ratingCount,omitempty"`
	RatingValue float64 `json:"ratingValue,omitempty"`
	BestRating  int     `json:"bestRating,omitempty"`
	WorstRating int     `json:"worstRating,omitempty"`
}

// ImageObject represents an image object https://schema.org/ImageObject
type ImageObject struct {
	Url     string `json:"url,omitempty"`
	Width   int    `json:"width,omitempty"`
	Height  int    `json:"height,omitempty"`
	Caption string `json:"caption,omitempty"`
}

// VideoObject represents a video object https://schema.org/VideoObject
type VideoObject struct {
	Name         string     `json:"name,omitempty"`
	Description  string     `json:"description,omitempty"`
	Duration     string     `json:"duration,omitempty"`
	EmbedUrl     string     `json:"embedUrl,omitempty"`
	ContentUrl   string     `json:"contentUrl,omitempty"`
	ThumbnailUrl string     `json:"thumbnailUrl,omitempty"`
	UploadDate   *time.Time `json:"uploadDate,omitempty"`
}

// Recipe is the basic struct for the recipe https://schema.org/Recipe
type Recipe struct {
	Url           string                `json:"url,omitempty"`
	Name          string                `json:"name,omitempty"`
	Description   string                `json:"description,omitempty"`
	Language      string                `json:"inLanguage,omitempty"`
	ThumbnailUrl  string                `json:"thumbnailUrl,omitempty"`
	Images        []*ImageObject        `json:"image,omitempty"`
	Author        *Person               `json:"author,omitempty"`
	Text          string                `json:"text,omitempty"`
	PrepTime      float64               `json:"prepTime,omitempty"`
	CookTime      float64               `json:"cookTime,omitempty"`
	TotalTime     float64               `json:"totalTime,omitempty"`
	Difficulty    string                `json:"difficulty,omitempty"` // alias `educationalLevel` TODO: difficulty is not a part of Recipe schema https://github.com/schemaorg/schemaorg/issues/3130
	CookingMethod string                `json:"cookingMethod,omitempty"`
	Diets         []string              `json:"suitableForDiet,omitempty"`
	Categories    []string              `json:"recipeCategory,omitempty"`
	Cuisines      []string              `json:"recipeCuisine,omitempty"`
	Keywords      []string              `json:"keywords,omitempty"`
	Yield         int                   `json:"recipeYield,omitempty"`        // alias `yield`
	Ingredients   []string              `json:"recipeIngredient,omitempty"`   // alias `supply`
	Equipment     []string              `json:"recipeEquipment,omitempty"`    // alias `tool` TODO: recipeEquipment is not a part of Recipe schema https://github.com/schemaorg/schemaorg/issues/3132
	Instructions  []*HowToSection       `json:"recipeInstructions,omitempty"` // alias `step`
	Notes         []string              `json:"correction,omitempty"`
	Nutrition     *NutritionInformation `json:"nutrition,omitempty"`
	Rating        *AggregateRating      `json:"aggregateRating,omitempty"`
	CommentCount  int                   `json:"commentCount,omitempty"`
	Video         *VideoObject          `json:"video,omitempty"`
	Links         []string              `json:"citation,omitempty"`
	Publisher     *Organization         `json:"publisher,omitempty"`
	DateModified  *time.Time            `json:"dateModified,omitempty"`
	DatePublished *time.Time            `json:"datePublished,omitempty"`
}

func (r *Recipe) AddImage(image *ImageObject) {
	if r.ThumbnailUrl == image.Url {
		return // don't add the same image twice
	}

	for _, vs := range r.Images { // check if already exists
		if image.Url == vs.Url {
			if image.Width > 0 {
				vs.Width = image.Width
			}
			if image.Height > 0 {
				vs.Height = image.Height
			}
			if len(image.Caption) != 0 {
				vs.Caption = image.Caption
			}
			return
		}
	}

	r.Images = append(r.Images, image)
}

func (r *Recipe) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "Unable to output in json: " + err.Error()
	}
	return string(data)
}