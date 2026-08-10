package custom

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/scraper/schema"
	"github.com/borschtapp/krip/utils"
)

type recipeNextData struct {
	Props struct {
		PageProps struct {
			SsrPayload struct {
				Recipe struct {
					RecipeID            string `json:"recipeId"`
					Name                string `json:"name"`
					ClonedFrom          string `json:"clonedFrom"`
					Description         string `json:"description"`
					DescriptionHTML     string `json:"descriptionHTML"`
					DescriptionMarkdown string `json:"descriptionMarkdown"`
					UUID                string `json:"uuid"`
					Label               []struct {
						ShowToCustomer  bool   `json:"showToCustomer"`
						ID              string `json:"id"`
						Type            string `json:"type"`
						Name            string `json:"name"`
						ForegroundColor string `json:"foregroundColor"`
						BackgroundColor string `json:"backgroundColor"`
					} `json:"label"`
					Allergens []struct {
						Slug             string `json:"slug"`
						IconPath         string `json:"iconPath"`
						TracesOf         bool   `json:"tracesOf"`
						TriggersTracesOf bool   `json:"triggersTracesOf"`
						ID               string `json:"id"`
						Type             string `json:"type"`
						Name             string `json:"name"`
					} `json:"allergens"`
					LanguageCode     string  `json:"languageCode"`
					ServingSize      int     `json:"servingSize"`
					Headline         string  `json:"headline"`
					Country          string  `json:"country"`
					UniqueRecipeCode string  `json:"uniqueRecipeCode"`
					ImagePath        string  `json:"imagePath"`
					TotalTime        string  `json:"totalTime"`
					PrepTime         string  `json:"prepTime"`
					Difficulty       int     `json:"difficulty"`
					IsAddon          bool    `json:"isAddon"`
					SeoName          string  `json:"seoName"`
					SeoDescription   string  `json:"seoDescription"`
					Canonical        string  `json:"canonical"`
					AverageRating    float64 `json:"averageRating"`
					FavoritesCount   int     `json:"favoritesCount"`
					RatingsCount     int     `json:"ratingsCount"`
					Nutrition        []struct {
						ID     string  `json:"id"`
						Type   string  `json:"type"`
						Name   string  `json:"name"`
						Unit   string  `json:"unit"`
						Amount float64 `json:"amount"`
					} `json:"nutrition"`
					Ingredients []struct {
						ID           string `json:"id"`
						Type         string `json:"type"`
						Shipped      bool   `json:"shipped"`
						ImagePath    string `json:"imagePath"`
						FamilyID     string `json:"familyId"`
						AllergensNew []struct {
							Name             string `json:"name"`
							Slug             string `json:"slug"`
							IconPath         string `json:"iconPath"`
							TracesOf         bool   `json:"tracesOf"`
							TriggersTracesOf bool   `json:"triggersTracesOf"`
							ID               string `json:"id"`
							Type             string `json:"type"`
						} `json:"allergensNew"`
						UUID      string `json:"uuid"`
						Name      string `json:"name"`
						Slug      string `json:"slug"`
						Allergens []struct {
							Name             string `json:"name"`
							Slug             string `json:"slug"`
							IconPath         string `json:"iconPath"`
							TracesOf         bool   `json:"tracesOf"`
							TriggersTracesOf bool   `json:"triggersTracesOf"`
							ID               string `json:"id"`
							Type             string `json:"type"`
						} `json:"allergens"`
					} `json:"ingredients"`
					Yields []struct {
						ID          string `json:"id"`
						Yields      int    `json:"yields"`
						Ingredients []struct {
							ID     string  `json:"id"`
							Amount float64 `json:"amount"`
							Unit   string  `json:"unit"`
						} `json:"ingredients"`
					} `json:"yields"`
					Steps []struct {
						InstructionsHTML     string `json:"instructionsHTML"`
						InstructionsMarkdown string `json:"instructionsMarkdown"`
						Images               []struct {
							Caption string `json:"caption"`
							Path    string `json:"path"`
							Link    string `json:"link"`
							ID      string `json:"id"`
						} `json:"images"`
						Videos       []interface{} `json:"videos"`
						ID           string        `json:"id"`
						Index        int           `json:"index"`
						Instructions string        `json:"instructions"`
					} `json:"steps"`
					Active bool   `json:"active"`
					Slug   string `json:"slug"`
					Tags   []struct {
						Slug         string      `json:"slug"`
						ColorHandle  string      `json:"colorHandle"`
						Preferences  interface{} `json:"preferences"`
						DisplayLabel bool        `json:"displayLabel"`
						ID           string      `json:"id"`
						Name         string      `json:"name"`
						Type         string      `json:"type"`
					} `json:"tags"`
					IsPublished bool `json:"isPublished"`
					Utensils    []struct {
						ID   string `json:"id"`
						Type string `json:"type"`
						Name string `json:"name"`
					} `json:"utensils"`
					Cuisines     json.RawMessage `json:"cuisines"`
					CreatedAt    time.Time       `json:"createdAt"`
					AllergensNew []struct {
						Slug             string `json:"slug"`
						IconPath         string `json:"iconPath"`
						TracesOf         bool   `json:"tracesOf"`
						TriggersTracesOf bool   `json:"triggersTracesOf"`
						ID               string `json:"id"`
						Type             string `json:"type"`
						Name             string `json:"name"`
					} `json:"allergensNew"`
					AdjustedRating           float64     `json:"adjustedRating"`
					CardLink                 string      `json:"cardLink"`
					Tier                     int         `json:"tier"`
					UpdatedCanonical         string      `json:"updatedCanonical"`
					UpdatedAt                time.Time   `json:"updatedAt"`
					IsCanonical              bool        `json:"isCanonical"`
					ID                       string      `json:"id"`
					Reviews                  interface{} `json:"reviews"`
					CanonicalLink            string      `json:"canonicalLink"`
					WebsiteURL               string      `json:"websiteUrl"`
					DisableTierAutoUpdate    bool        `json:"disableTierAutoUpdate"`
					ReviewSummaryHighlights  interface{} `json:"reviewSummaryHighlights"`
					AggregateRating          float64     `json:"aggregateRating"`
					AggregateRatingsCount    int         `json:"aggregateRatingsCount"`
					ReviewSummaryLastUpdated interface{} `json:"reviewSummaryLastUpdated"`
				} `json:"recipe"`
				RecipeSupportedLocales []string `json:"recipeSupportedLocales"`
				RecipeSlugMap          struct {
					DeDE string `json:"de-DE"`
				} `json:"recipeSlugMap"`
				ContentfulEntries struct {
					Fields struct {
						ID          string `json:"id"`
						Locale      string `json:"locale"`
						Collections []struct {
							URL         string   `json:"url"`
							Name        string   `json:"name"`
							Ingredients []string `json:"ingredients"`
						} `json:"collections"`
						CuisineCollections []struct {
							URL     string `json:"url"`
							Name    string `json:"name"`
							Cuisine string `json:"cuisine"`
						} `json:"cuisineCollections"`
					} `json:"fields"`
					Sections  []interface{} `json:"sections"`
					CreatedAt time.Time     `json:"createdAt"`
					UpdatedAt time.Time     `json:"updatedAt"`
				} `json:"contentfulEntries"`
			} `json:"ssrPayload"`
		} `json:"pageProps"`
	} `json:"props"`
}

func ScrapeHelloFresh(data *model.DataInput, r *model.Recipe) error {
	if err := schema.Scrape(data, r); err != nil {
		return err
	}

	if data.Document == nil {
		return nil
	}

	nextDataRaw := data.Document.Find("script#__NEXT_DATA__").Text()
	if nextDataRaw == "" {
		return nil
	}

	var nextDataObj recipeNextData
	if err := json.Unmarshal([]byte(nextDataRaw), &nextDataObj); err != nil {
		return fmt.Errorf("json unmarshal error in HelloFresh nextData: %w", err)
	}

	recipe := nextDataObj.Props.PageProps.SsrPayload.Recipe
	imgBaseUrl := "https://media.hellofresh.com"
	imgRecipePrefix := "/w_1200,q_auto,f_auto,c_limit,fl_lossy/hellofresh_s3"
	imgStepPrefix := "/w_750,q_auto,f_auto,c_limit,fl_lossy/hellofresh_s3"
	imgIngredientPrefix := "/w_256,q_auto,f_auto,c_limit,fl_lossy/hellofresh_s3"

	if recipe.ImagePath != "" {
		r.Images = make([]*model.ImageObject, 0, 1)
		r.AddImageUrl(imgBaseUrl + imgRecipePrefix + recipe.ImagePath)
	}

	if recipe.LanguageCode != "" && r.Language == "" {
		r.Language = recipe.LanguageCode
	}

	if recipe.Difficulty > 0 && r.Difficulty == "" {
		r.Difficulty = strconv.Itoa(recipe.Difficulty)
	}

	if len(recipe.Label) > 0 {
		for _, l := range recipe.Label {
			if l.ShowToCustomer && l.Name != "" {
				r.Keywords = append(r.Keywords, l.Name)
			}
		}
	}

	if len(recipe.Yields) > 0 && r.Yield == "" {
		r.Yield = fmt.Sprint(recipe.Yields[0].Yields)
	}

	if len(recipe.Tags) > 0 && len(r.Categories) == 0 {
		for _, tag := range recipe.Tags {
			if tag.DisplayLabel && tag.Name != "" {
				r.Categories = append(r.Categories, tag.Name)
			}
		}
	}

	if len(recipe.Utensils) > 0 {
		r.Equipment = make([]*model.HowToTool, 0, len(recipe.Utensils))
		for _, u := range recipe.Utensils {
			r.Equipment = append(r.Equipment, &model.HowToTool{Name: u.Name})
		}
	}

	if recipe.AggregateRating > 0 && r.Rating == nil {
		r.Rating = &model.AggregateRating{
			RatingValue: recipe.AggregateRating,
			RatingCount: recipe.AggregateRatingsCount,
		}
	}

	if len(recipe.Nutrition) > 0 {
		if r.Nutrition == nil {
			r.Nutrition = &model.NutritionInformation{}
		}
		if recipe.ServingSize > 0 {
			r.Nutrition.ServingSize = fmt.Sprintf("%d g", recipe.ServingSize)
		}
		nutritionMap := map[string]**float64{
			"57b42a48b7e8697d4b305304": &r.Nutrition.Calories,            // calories (kcal)
			"57b42a48b7e8697d4b305307": &r.Nutrition.FatContent,          // fat
			"57b42a48b7e8697d4b305308": &r.Nutrition.SaturatedFatContent, // saturated fat
			"57b42a48b7e8697d4b305305": &r.Nutrition.CarbohydrateContent, // carbohydrates
			"57b42a48b7e8697d4b305306": &r.Nutrition.SugarContent,        // sugar
			"57b42a48b7e8697d4b30530a": &r.Nutrition.FiberContent,        // fiber
			"57b42a48b7e8697d4b305309": &r.Nutrition.ProteinContent,      // protein
			"57b42a48b7e8697d4b30530b": &r.Nutrition.SaltContent,         // salt (g) — HelloFresh mislabels Salt as sodiumContent in JSON-LD
			"652d4d58ce1a3c29bd168c82": &r.Nutrition.PotassiumContent,    // potassium (mg)
			"652d4d6bce1a3c29bd168c84": &r.Nutrition.CalciumContent,      // calcium (mg)
			"652d4d7bce1a3c29bd168c86": &r.Nutrition.IronContent,         // iron (mg)
		}
		for _, n := range recipe.Nutrition {
			val := n.Amount
			if field, ok := nutritionMap[n.Type]; ok {
				*field = &val
			}
		}
	}

	if recipe.CardLink != "" {
		r.Links = append(r.Links, recipe.CardLink)
	}

	if !recipe.CreatedAt.IsZero() && r.DatePublished == nil {
		r.DatePublished = &recipe.CreatedAt
	}
	if !recipe.UpdatedAt.IsZero() && r.DateModified == nil {
		r.DateModified = &recipe.UpdatedAt
	}

	if len(recipe.Steps) > 0 {
		r.Instructions = make([]*model.HowToSection, 0, len(recipe.Steps))
		for _, step := range recipe.Steps {
			howToSection := &model.HowToSection{}
			howToSection.Text = utils.CleanupInline(step.Instructions)

			if len(step.Images) > 0 && step.Images[0].Path != "" {
				howToSection.Image = imgBaseUrl + imgStepPrefix + step.Images[0].Path
			}

			r.Instructions = append(r.Instructions, howToSection)
		}
	}

	if len(recipe.Cuisines) > 0 && len(r.Cuisines) == 0 {
		// cuisines can be a plain string, []string, or []{"name":...} objects
		var strVal string
		var strSlice []string
		var objSlice []struct {
			Name string `json:"name"`
		}
		switch {
		case json.Unmarshal(recipe.Cuisines, &strVal) == nil:
			r.Cuisines = []string{strVal}
		case json.Unmarshal(recipe.Cuisines, &strSlice) == nil:
			r.Cuisines = strSlice
		case json.Unmarshal(recipe.Cuisines, &objSlice) == nil:
			for _, c := range objSlice {
				if c.Name != "" {
					r.Cuisines = append(r.Cuisines, c.Name)
				}
			}
		}
	}

	if len(recipe.Ingredients) > 0 {
		r.Ingredients = make([]*model.PropertyValue, 0, len(recipe.Ingredients))
		for _, ing := range recipe.Ingredients {
			propValue := &model.PropertyValue{}
			propValue.Name = utils.CleanupInline(ing.Name)
			propValue.Category = ing.Type
			propValue.Pantry = !ing.Shipped

			if ing.ImagePath != "" {
				propValue.Image = imgBaseUrl + imgIngredientPrefix + ing.ImagePath
			}

			if len(recipe.Yields) > 0 {
				for _, yieldIng := range recipe.Yields[0].Ingredients {
					if yieldIng.ID == ing.ID {
						propValue.Value = strconv.FormatFloat(yieldIng.Amount, 'f', -1, 64)
						propValue.UnitText = yieldIng.Unit
						if yieldIng.Amount > 0 && yieldIng.Unit != "" {
							propValue.Description = propValue.Value + " " + yieldIng.Unit + " " + propValue.Name
						} else if yieldIng.Amount > 0 {
							propValue.Description = propValue.Value + " " + propValue.Name
						}
						break
					}
				}
			}

			for _, a := range ing.AllergensNew {
				if a.TriggersTracesOf {
					continue // skip meta-allergen entries
				}
				propValue.Allergens = append(propValue.Allergens, &model.Allergen{
					Name:     a.Name,
					TracesOf: a.TracesOf,
				})
			}

			r.Ingredients = append(r.Ingredients, propValue)
		}
	}

	if len(recipe.AllergensNew) > 0 {
		for _, a := range recipe.AllergensNew {
			if a.TriggersTracesOf {
				continue
			}
			r.Allergens = append(r.Allergens, &model.Allergen{
				Name:     a.Name,
				TracesOf: a.TracesOf,
			})
		}
	}

	return nil
}

type feedNextData struct {
	Props struct {
		PageProps struct {
			SsrPayload struct {
				ActiveWeek string `json:"activeWeek"`
				Config     struct {
					HeadMeta struct {
						BrandName string `json:"brandName"`
					} `json:"head-meta-tags-feature-config"`
					Header struct {
						Logo struct {
							LogoURL string `json:"logoURL"`
						} `json:"logo"`
					} `json:"hf.funnel.header"`
				} `json:"config"`
				Courses []struct {
					Index  int `json:"index"`
					Recipe struct {
						Country  string `json:"country"`
						Cuisines []struct {
							ID       string `json:"id"`
							Type     string `json:"type"`
							Name     string `json:"name"`
							Slug     string `json:"slug"`
							IconLink string `json:"iconLink"`
						} `json:"cuisines"`
						Difficulty     int    `json:"difficulty"`
						FavoritesCount int    `json:"favoritesCount"`
						Headline       string `json:"headline"`
						ID             string `json:"id"`
						ImageLink      string `json:"imageLink"`
						ImagePath      string `json:"imagePath"`
						IsPublished    bool   `json:"isPublished"`
						Label          struct {
							Text   string `json:"text"`
							Handle string `json:"handle"`
						} `json:"label"`
						Name         string `json:"name"`
						PrepTime     string `json:"prepTime"`
						RatingsCount int    `json:"ratingsCount"`
						Slug         string `json:"slug"`
						Tags         []struct {
							ID   string `json:"id"`
							Type string `json:"type"`
							Name string `json:"name"`
							Slug string `json:"slug"`
						} `json:"tags"`
						TotalTime  string `json:"totalTime"`
						UUID       string `json:"uuid"`
						WebsiteURL string `json:"websiteUrl"`
					} `json:"recipe"`
				} `json:"courses"`
			} `json:"ssrPayload"`
		} `json:"pageProps"`
	} `json:"props"`
}

func ScrapeHelloFreshFeed(data *model.DataInput, feed *model.Feed) error {
	if data.Document == nil {
		return fmt.Errorf("no document found")
	}

	baseUrl, err := url.Parse(data.Url)
	if err != nil {
		return err
	}

	// Verify that the page is a menus page (e.g., /menus, /menus/, or regional variants)
	if !strings.HasPrefix(baseUrl.Path, "/menus") {
		return fmt.Errorf("not a menus page")
	}

	nextDataRaw := data.Document.Find("script#__NEXT_DATA__").Text()
	if nextDataRaw == "" {
		return fmt.Errorf("no next data found")
	}

	var nextDataObj feedNextData
	if err := json.Unmarshal([]byte(nextDataRaw), &nextDataObj); err != nil {
		return fmt.Errorf("json unmarshal error: %w", err)
	}

	payload := nextDataObj.Props.PageProps.SsrPayload

	var publisher *model.Organization
	if payload.Config.HeadMeta.BrandName != "" {
		publisher = &model.Organization{
			Name: payload.Config.HeadMeta.BrandName,
			Logo: payload.Config.Header.Logo.LogoURL,
			Url:  utils.BaseUrl(data.Url),
		}
	}

	var weekDate *time.Time
	if payload.ActiveWeek != "" {
		var year, week int
		if n, _ := fmt.Sscanf(payload.ActiveWeek, "%d-W%d", &year, &week); n == 2 {
			// ISO week 1 always contains Jan 4.
			jan4 := time.Date(year, 1, 4, 0, 0, 0, 0, time.UTC)
			// Find the Monday of that week
			daysToSubtract := (int(jan4.Weekday()) + 6) % 7
			mondayOfW1 := jan4.AddDate(0, 0, -daysToSubtract)
			// Add (week - 1) * 7 days to get the Monday of the target week
			t := mondayOfW1.AddDate(0, 0, (week-1)*7)
			weekDate = &t
		}
	}

	for _, course := range payload.Courses {
		if course.Recipe.Name != "" {
			var categories []string
			var cuisines []string
			for _, t := range course.Recipe.Tags {
				categories = append(categories, t.Name)
			}
			for _, c := range course.Recipe.Cuisines {
				cuisines = append(cuisines, c.Name)
			}

			entry := &model.Recipe{
				Name:          course.Recipe.Name,
				Url:           course.Recipe.WebsiteURL,
				Description:   course.Recipe.Headline,
				Publisher:     publisher,
				TotalTime:     course.Recipe.TotalTime,
				Cuisines:      cuisines,
				Categories:    categories,
				Difficulty:    fmt.Sprint(course.Recipe.Difficulty),
				DatePublished: weekDate,
			}
			entry.AddImageUrl(course.Recipe.ImageLink)
			feed.AddEntry(entry)
		}
	}

	return nil
}
