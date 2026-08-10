package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Person according to https://schema.org/Person
type Person struct {
	Name        string   `json:"name,omitempty"`
	JobTitle    string   `json:"jobTitle,omitempty"`
	Description string   `json:"description,omitempty"`
	KnowsAbout  []string `json:"knowsAbout,omitempty"`
	Url         string   `json:"url,omitempty"`
	Image       string   `json:"image,omitempty"`
}

// Organization according to https://schema.org/Organization
type Organization struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Url         string `json:"url,omitempty"`
	Logo        string `json:"logo,omitempty"`
}

type Allergen struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	TracesOf    bool   `json:"tracesOf,omitempty"`
}

// PropertyValue represents a property-value pair, e.g. an ingredient and its amount https://schema.org/PropertyValue
type PropertyValue struct {
	Value       string `json:"value,omitempty"` // The quantitative value of the property, e.g. "2", "1/2", "a pinch"
	MaxValue    string `json:"maxValue,omitempty"`
	MinValue    string `json:"minValue,omitempty"`
	UnitText    string `json:"unitText,omitempty"` // The unit of measurement, e.g. "g", "cup", "teaspoon"
	UnitCode    string `json:"unitCode,omitempty"`
	Name        string `json:"name,omitempty"` // The name of the property, e.g. "sugar", "flour", "salt"
	Image       string `json:"image,omitempty"`
	Url         string `json:"url,omitempty"`
	Description string `json:"description,omitempty"`
	// extra fields not covered by schema.org
	Category      string      `json:"category,omitempty"`
	Pantry        bool        `json:"pantry,omitempty"`
	EstimatedCost string      `json:"estimatedCost,omitempty"`
	Allergens     []*Allergen `json:"allergens,omitempty"`
}

// HowToTool represents a tool used in the instructions https://schema.org/HowToTool
type HowToTool struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Url         string `json:"url,omitempty"`
	Image       string `json:"image,omitempty"`
	Quantity    string `json:"requiredQuantity,omitempty"`
}

// HowToStep a step in the instructions https://schema.org/HowToStep
type HowToStep struct {
	Name  string `json:"name,omitempty"`
	Text  string `json:"text,omitempty"`
	Url   string `json:"url,omitempty"`
	Image string `json:"image,omitempty"`
	Video string `json:"video,omitempty"`
}

// HowToSection a group of steps in the instructions https://schema.org/HowToSection
type HowToSection struct {
	HowToStep              // because it's optional to have a group, we have to embed `HowToStep` here
	Steps     []*HowToStep `json:"itemListElement,omitempty"`
}

// NutritionInformation according to https://schema.org/NutritionInformation
type NutritionInformation struct {
	ServingSize           string   `json:"servingSize,omitempty"`           // The serving size, in terms of the number of volume or mass.
	Calories              *float64 `json:"calories,omitempty"`              // The number of calories.
	CarbohydrateContent   *float64 `json:"carbohydrateContent,omitempty"`   // The number of grams of carbohydrates.
	CholesterolContent    *float64 `json:"cholesterolContent,omitempty"`    // The number of milligrams of cholesterol.
	FatContent            *float64 `json:"fatContent,omitempty"`            // The number of grams of fat.
	FiberContent          *float64 `json:"fiberContent,omitempty"`          // The number of grams of fiber.
	ProteinContent        *float64 `json:"proteinContent,omitempty"`        // The number of grams of protein.
	SaturatedFatContent   *float64 `json:"saturatedFatContent,omitempty"`   // The number of grams of saturated fat.
	SodiumContent         *float64 `json:"sodiumContent,omitempty"`         // The number of milligrams of sodium.
	SugarContent          *float64 `json:"sugarContent,omitempty"`          // The number of grams of sugar.
	TransFatContent       *float64 `json:"transFatContent,omitempty"`       // The number of grams of trans fat.
	UnsaturatedFatContent *float64 `json:"unsaturatedFatContent,omitempty"` // The number of grams of unsaturated fat.
	// other minerals commonly found in recipes, not covered by schema.org
	SaltContent      *float64 `json:"saltContent,omitempty"`      // The number of grams of salt.
	IronContent      *float64 `json:"ironContent,omitempty"`      // The number of milligrams of iron.
	PotassiumContent *float64 `json:"potassiumContent,omitempty"` // The number of milligrams of potassium.
	CalciumContent   *float64 `json:"calciumContent,omitempty"`   // The number of milligrams of calcium.
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
// Perhaps, I would rename recipeYield, recipeIngredient, recipeInstructions to their aliases,
// but many websites expect only these names (like Google Search https://developers.google.com/search/docs/appearance/structured-data/recipe)
type Recipe struct {
	Url           string                `json:"url,omitempty"`
	Name          string                `json:"name,omitempty"`
	Description   string                `json:"description,omitempty"`
	Language      string                `json:"inLanguage,omitempty"`
	Images        []*ImageObject        `json:"image,omitempty"`
	Author        *Person               `json:"author,omitempty"`
	Publisher     *Organization         `json:"publisher,omitempty"`
	Text          string                `json:"text,omitempty"`
	PrepTime      string                `json:"prepTime,omitempty"`
	CookTime      string                `json:"cookTime,omitempty"` // alias `performTime`
	TotalTime     string                `json:"totalTime,omitempty"`
	Difficulty    string                `json:"educationalLevel,omitempty"` // `difficulty` is not a part of Recipe schema https://github.com/schemaorg/schemaorg/issues/3130
	CookingMethod string                `json:"cookingMethod,omitempty"`
	Diets         []string              `json:"suitableForDiet,omitempty"`
	Categories    []string              `json:"recipeCategory,omitempty"`
	Cuisines      []string              `json:"recipeCuisine,omitempty"`
	Keywords      []string              `json:"keywords,omitempty"`
	Yield         string                `json:"recipeYield,omitempty"`        // alias `yield`
	Ingredients   []*PropertyValue      `json:"recipeIngredient,omitempty"`   // alias `supply`
	Equipment     []*HowToTool          `json:"tool,omitempty"`               // alias `tool`, `recipeEquipment` is not a part of Recipe schema https://github.com/schemaorg/schemaorg/issues/3132
	Instructions  []*HowToSection       `json:"recipeInstructions,omitempty"` // alias `step`
	Nutrition     *NutritionInformation `json:"nutrition,omitempty"`
	Rating        *AggregateRating      `json:"aggregateRating,omitempty"`
	CommentCount  int                   `json:"commentCount,omitempty"`
	Video         *VideoObject          `json:"video,omitempty"`
	Links         []string              `json:"sameAs,omitempty"` // maybe not the cleanest name, but we can store additional links here
	EstimatedCost string                `json:"estimatedCost,omitempty"`
	Allergens     []*Allergen           `json:"allergens,omitempty"` // not a part of Recipe schema
	DateModified  *time.Time            `json:"dateModified,omitempty"`
	DatePublished *time.Time            `json:"datePublished,omitempty"`

	// Scraped marks that this recipe went through a full Scrape pass, as opposed to being
	// a stub populated only from a feed listing (title/image/URL).
	Scraped bool `json:"-"`
}

func (r *Recipe) AddImageUrl(imageUrl string) {
	r.AddImage(&ImageObject{Url: imageUrl})
}

func (r *Recipe) AddImage(image *ImageObject) {
	if len(image.Url) == 0 {
		return
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

func (r *Recipe) IsValid() bool {
	return r.Validate(RecipeFilter{}) == nil
}

var (
	ErrNoNameOrUrl       = errors.New("recipe has no name or url")
	ErrNoImages          = errors.New("recipe has no images")
	ErrNoPublisher       = errors.New("recipe has no publisher")
	ErrNoIngredients     = errors.New("recipe has no ingredients")
	ErrTooFewIngredients = errors.New("recipe has too few ingredients")
	ErrNoInstructions    = errors.New("recipe has no instructions")
)

func (r *Recipe) Validate(filter RecipeFilter) error {
	if len(r.Name) == 0 || len(r.Url) == 0 {
		return ErrNoNameOrUrl
	}

	if !filter.OptionalImage && len(r.Images) == 0 {
		return ErrNoImages
	}

	if !filter.OptionalPublisher && (r.Publisher == nil || len(r.Publisher.Name) == 0) {
		return ErrNoPublisher
	}

	if !filter.OptionalIngredients && len(r.Ingredients) == 0 {
		return ErrNoIngredients
	}

	if filter.MinIngredients > 0 && len(r.Ingredients) < filter.MinIngredients {
		return fmt.Errorf("%w (%d < %d)", ErrTooFewIngredients, len(r.Ingredients), filter.MinIngredients)
	}

	if !filter.OptionalInstructions && len(r.Instructions) == 0 {
		return ErrNoInstructions
	}

	return nil
}

func (r *Recipe) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "Unable to output in json: " + err.Error()
	}
	return string(data)
}
