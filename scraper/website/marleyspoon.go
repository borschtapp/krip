package website

import (
	"encoding/json"
	"errors"
	"github.com/sosodev/duration"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/utils"
)

var idPattern = regexp.MustCompile(`/(\d+)-`)
var scriptPattern = regexp.MustCompile(`gon\.current_brand="([^"]+?)".*?gon\.current_country="([^"]+?)".*?gon\.api_token="([^"]+?)".*?gon\.api_host="([^"]+?)".*?`)

// these values I was able to retrieve from website
var preparationMap = map[string]time.Duration{
	"time_level_1": 10 * time.Minute, // on the website they are displayed like `5-10 minutes`, I used avg or similar rounded value
	"time_level_2": 15 * time.Minute,
	"time_level_3": 20 * time.Minute,
	"time_level_4": 25 * time.Minute,
	"time_level_5": 35 * time.Minute,
}

// MarleySpoonData struct is generated using https://mholt.github.io/json-to-go/
type MarleySpoonData struct {
	ID               int      `json:"id,omitempty"`
	Name             string   `json:"name,omitempty"`
	Subtitle         string   `json:"subtitle,omitempty"`
	NameWithSubtitle string   `json:"name_with_subtitle,omitempty"`
	Classic          bool     `json:"classic,omitempty"`
	Slug             string   `json:"slug,omitempty"`
	VariantID        int      `json:"variant_id,omitempty"`
	Country          string   `json:"country,omitempty"`
	Brand            string   `json:"brand,omitempty"`
	Description      string   `json:"description,omitempty"`
	MealType         string   `json:"meal_type,omitempty"`
	Calories         int      `json:"calories,omitempty"`
	Difficulty       string   `json:"difficulty,omitempty"`
	PreparationTime  string   `json:"preparation_time,omitempty"`
	ProductType      string   `json:"product_type,omitempty"`
	MealAttributes   []string `json:"meal_attributes,omitempty"`
	Nutrition        struct {
		Calories string `json:"calories,omitempty"`
		Carbs    string `json:"carbs,omitempty"`
		Proteins string `json:"proteins,omitempty"`
		Fat      string `json:"fat,omitempty"`
	} `json:"nutrition,omitempty"`
	Sku           string `json:"sku,omitempty"`
	RecipeCardURL string `json:"recipe_card_url,omitempty"`
	Image         struct {
		Thumbnail string `json:"thumbnail,omitempty"`
		Small     string `json:"small,omitempty"`
		Medium    string `json:"medium,omitempty"`
		Large     string `json:"large,omitempty"`
	} `json:"image,omitempty"`
	AdditionalAllergens []string `json:"additional_allergens,omitempty"`
	Steps               []struct {
		Position    int    `json:"position,omitempty"`
		Title       string `json:"title,omitempty"`
		Description string `json:"description,omitempty"`
		Photo       string `json:"photo,omitempty"`
	} `json:"steps,omitempty"`
	Ingredients []struct {
		Name  string `json:"name,omitempty"`
		Image struct {
			Thumbnail string `json:"thumbnail,omitempty"`
			Medium    string `json:"medium,omitempty"`
		} `json:"image,omitempty"`
		Allergens        []string `json:"allergens,omitempty"`
		NameWithQuantity string   `json:"name_with_quantity,omitempty"`
	} `json:"ingredients,omitempty"`
	AssumedIngredients []struct {
		Name string `json:"name,omitempty"`
	} `json:"assumed_ingredients,omitempty"`
	AssumedCookingUtilities []struct {
		Name string `json:"name,omitempty"`
	} `json:"assumed_cooking_utilities,omitempty"`
	Chef struct {
		Name        string `json:"name,omitempty"`
		Description string `json:"description,omitempty"`
		Bio         string `json:"bio,omitempty"`
		Image       struct {
			Thumbnail string `json:"thumbnail,omitempty"`
			Medium    string `json:"medium,omitempty"`
		} `json:"image,omitempty"`
		Slug string `json:"slug,omitempty"`
	} `json:"chef,omitempty"`
	CookingTip interface{} `json:"cooking_tip,omitempty"`
}

func ScrapeMarleySpoon(data *model.DataInput, r *model.Recipe) error {
	if data.Document != nil {
		apiUrl, apiToken, err := findApiParams(data.Document, data.Url)
		if err != nil {
			return err
		}

		body, _, err := utils.ReadUrl(apiUrl, map[string][]string{
			"Accept":        {"application/json"},
			"Authorization": {apiToken},
		})
		if err != nil {
			return err
		}

		data := MarleySpoonData{}
		if err := json.Unmarshal(body, &data); err != nil {
			return err
		}

		if err := parseData(&data, r); err != nil {
			return err
		}
	}

	return nil
}

func parseData(data *MarleySpoonData, r *model.Recipe) error {
	if len(data.NameWithSubtitle) != 0 {
		r.Name = utils.CleanupInline(data.NameWithSubtitle)
	}

	if len(data.PreparationTime) != 0 {
		r.TotalTime = duration.Format(preparationMap[data.PreparationTime])
	}

	// The backend of MarleySpoon always returns ingredients for 2 servings
	// This conclusion is made based on personal observations and available plans https://marleyspoon.com/select-plan
	r.Yield = 2

	if len(data.Difficulty) != 0 {
		r.Difficulty = data.Difficulty
	}

	if len(data.MealType) != 0 {
		for _, diet := range strings.Split(data.MealType, ",") {
			r.Diets = utils.AppendUnique(r.Diets, utils.CleanupInline(diet))
		}
	}

	if len(data.Image.Thumbnail) != 0 {
		r.AddImage(&model.ImageObject{Url: data.Image.Thumbnail})
	}

	if len(data.Image.Medium) != 0 {
		r.AddImage(&model.ImageObject{Url: data.Image.Medium})
	}

	if len(data.Image.Large) != 0 {
		r.AddImage(&model.ImageObject{Url: data.Image.Large})
	}

	if len(data.Nutrition.Calories) != 0 || len(data.Nutrition.Fat) != 0 || len(data.Nutrition.Carbs) != 0 || len(data.Nutrition.Proteins) != 0 {
		var nutrition model.NutritionInformation
		nutrition.Calories = data.Nutrition.Calories
		nutrition.FatContent = data.Nutrition.Fat
		nutrition.CarbohydrateContent = data.Nutrition.Carbs
		nutrition.ProteinContent = data.Nutrition.Proteins
		r.Nutrition = &nutrition
	}

	if len(data.Ingredients) != 0 {
		for _, ingredient := range data.Ingredients {
			if len(ingredient.NameWithQuantity) != 0 {
				r.Ingredients = append(r.Ingredients, ingredient.NameWithQuantity)
			} else {
				r.Ingredients = append(r.Ingredients, ingredient.Name)
			}
		}
	}

	if len(data.AssumedIngredients) != 0 {
		for _, ingredient := range data.AssumedIngredients {
			r.Ingredients = append(r.Ingredients, ingredient.Name)
		}
	}

	if len(data.AssumedCookingUtilities) != 0 {
		for _, utility := range data.AssumedCookingUtilities {
			r.Equipment = append(r.Equipment, utility.Name)
		}
	}

	if len(data.Steps) != 0 {
		for _, item := range data.Steps {
			var step model.HowToStep
			step.Name = utils.CleanupInline(item.Title)
			step.Text = utils.CleanupInline(strings.ReplaceAll(item.Description, "__", ""))
			step.Image = item.Photo
			r.Instructions = append(r.Instructions, &model.HowToSection{HowToStep: step})
		}
	}

	if len(data.MealAttributes) != 0 {
		for _, attr := range data.MealAttributes {
			r.Keywords = utils.AppendUnique(r.Keywords, utils.CleanupInline(strings.ReplaceAll(attr, "_", " ")))
		}
	}

	if len(data.Chef.Name) != 0 {
		var author model.Person
		author.Name = data.Chef.Name
		author.Description = data.Chef.Bio
		author.Image = data.Chef.Image.Medium
		r.Author = &author
	}

	if len(data.Description) != 0 {
		r.Description = utils.CleanupInline(data.Description)
	}

	if len(data.RecipeCardURL) != 0 {
		r.Links = append(r.Links, data.RecipeCardURL)
	}

	// in normal scenario, there will be html `lang` tag and language can be retrieved from it
	if len(r.Language) == 0 && len(data.Country) != 0 {
		// but using the `country` property, we can guess it
		r.Language = utils.CleanupLang(data.Country)
	}

	return nil
}

func findApiParams(doc *goquery.Document, canonicalUrl string) (url string, token string, err error) {
	if match := idPattern.FindStringSubmatch(canonicalUrl); match != nil {
		recipeId := match[1]

		doc.Find("script").EachWithBreak(func(i int, s *goquery.Selection) bool {
			script := s.Text()
			if match := scriptPattern.FindStringSubmatch(script); match != nil {
				host := strings.ReplaceAll(match[4], "\\", "")

				url = host + "/recipes/" + recipeId + "?brand=" + match[1] + "&country=" + match[2] + "&product_type=web"
				token = "Bearer " + match[3]
				return false
			}
			return true
		})

		if len(url) == 0 || len(token) == 0 {
			err = errors.New("could not find api params")
		}
	} else {
		err = errors.New("could not find recipe id")
	}

	return
}
