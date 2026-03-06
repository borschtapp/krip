package custom

import (
	"encoding/json"
	"fmt"
	"net/url"
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
					Cuisines []struct {
						Country   string `json:"country"`
						Locale    string `json:"locale"`
						Name      string `json:"name"`
						Slug      string `json:"slug"`
						Type      string `json:"type"`
						ID        string `json:"id"`
						CuisineID string `json:"cuisineId"`
					} `json:"cuisines"`
					CreatedAt    time.Time `json:"createdAt"`
					AllergensNew []struct {
						Slug             string `json:"slug"`
						IconPath         string `json:"iconPath"`
						TracesOf         bool   `json:"tracesOf"`
						TriggersTracesOf bool   `json:"triggersTracesOf"`
						ID               string `json:"id"`
						Type             string `json:"type"`
						Name             string `json:"name"`
					} `json:"allergensNew"`
					AdjustedRating   float64   `json:"adjustedRating"`
					CardLink         string    `json:"cardLink"`
					Tier             int       `json:"tier"`
					UpdatedCanonical string    `json:"updatedCanonical"`
					UpdatedAt        time.Time `json:"updatedAt"`
					HasBeenProcessed struct {
						QualityJudge     bool `json:"qualityJudge"`
						UpdatedCanonical bool `json:"updatedCanonical"`
						Tier             bool `json:"tier"`
					} `json:"hasBeenProcessed"`
					IsCanonical   bool   `json:"isCanonical"`
					ID            string `json:"id"`
					VideoMetadata struct {
						Thumbnails struct {
						} `json:"thumbnails"`
					} `json:"videoMetadata"`
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
		return fmt.Errorf("json unmarshal error in HelloFresh nextData: %v", err)
	}

	recipe := nextDataObj.Props.PageProps.SsrPayload.Recipe
	imgBaseUrl := "https://media.hellofresh.com"
	imgStepPrefix := "/w_750,q_auto,f_auto,c_limit,fl_lossy/hellofresh_s3"
	imgIngredientPrefix := "/w_256,q_auto,f_auto,c_limit,fl_lossy/hellofresh_s3"

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

	if len(recipe.Ingredients) > 0 {
		r.Ingredients = make([]*model.PropertyValue, 0, len(recipe.Ingredients))
		for _, ing := range recipe.Ingredients {
			propValue := &model.PropertyValue{}
			propValue.Name = utils.CleanupInline(ing.Name)

			if ing.ImagePath != "" {
				propValue.Image = imgBaseUrl + imgIngredientPrefix + ing.ImagePath
			}

			if len(recipe.Yields) > 0 {
				for _, yieldIng := range recipe.Yields[0].Ingredients {
					if yieldIng.ID == ing.ID {
						propValue.Value = fmt.Sprint(yieldIng.Amount)
						propValue.UnitText = yieldIng.Unit
						break
					}
				}
			}
			r.Ingredients = append(r.Ingredients, propValue)
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
		return fmt.Errorf("json unmarshal error: %v", err)
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
			// Rough estimation of the week's start date
			t := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
			t = t.AddDate(0, 0, (week-1)*7)
			weekDate = &t
		}
	}

	uniqueEntries := make(map[string]*model.Recipe)

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
			uniqueEntries[entry.Url] = entry
		}
	}

	for _, entry := range uniqueEntries {
		feed.Entries = append(feed.Entries, entry)
	}

	return nil
}
