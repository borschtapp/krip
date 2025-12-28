package website

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/utils"
	"github.com/sosodev/duration"
)

type GoustoData struct {
	Status string `json:"status"`
	Data   struct {
		Entry struct {
			URL        string `json:"url"`
			Title      string `json:"title"`
			Categories []struct {
				Title string `json:"title"`
				URL   string `json:"url"`
				UID   string `json:"uid"`
			} `json:"categories"`
			GoustoID  string `json:"gousto_id"`
			GoustoUID string `json:"gousto_uid"`
			Media     struct {
				Images []struct {
					Image string `json:"image"`
					Width int    `json:"width"`
				} `json:"images"`
			} `json:"media"`
			Rating struct {
				Average float64 `json:"average"`
				Count   int     `json:"count"`
			} `json:"rating"`
			Description string `json:"description"`
			PrepTimes   struct {
				For2 int `json:"for_2"`
				For4 int `json:"for_4"`
			} `json:"prep_times"`
			Cuisine struct {
				Slug  string `json:"slug"`
				Title string `json:"title"`
			} `json:"cuisine"`
			Ingredients []struct {
				Label string `json:"label"`
				Title string `json:"title"`
				UID   string `json:"uid"`
				Name  string `json:"name"`
				Media struct {
					Images []struct {
						Image string `json:"image"`
						Width int    `json:"width"`
					} `json:"images"`
				} `json:"media"`
				Allergens struct {
					Allergen []any `json:"allergen"`
				} `json:"allergens"`
			} `json:"ingredients"`
			Basics []struct {
				Title string `json:"title"`
				Slug  string `json:"slug"`
			} `json:"basics"`
			CookingInstructions []struct {
				Instruction string `json:"instruction"`
				Order       int    `json:"order"`
				Media       struct {
					Images []struct {
						Image string `json:"image"`
						Width int    `json:"width"`
					} `json:"images"`
				} `json:"media"`
			} `json:"cooking_instructions"`
			Allergens []struct {
				Title string `json:"title"`
				Slug  string `json:"slug"`
			} `json:"allergens"`
			Seo struct {
				Title          string `json:"title"`
				Description    string `json:"description"`
				Robots         []any  `json:"robots"`
				Canonical      string `json:"canonical"`
				OpenGraphImage string `json:"open_graph_image"`
			} `json:"seo"`
			Tags                   []any  `json:"tags"`
			UID                    string `json:"uid"`
			Version                int    `json:"_version"`
			NutritionalInformation struct {
				PerHundredGrams struct {
					EnergyKcal     int `json:"energy_kcal"`
					EnergyKj       int `json:"energy_kj"`
					FatMg          int `json:"fat_mg"`
					FatSaturatesMg int `json:"fat_saturates_mg"`
					CarbsMg        int `json:"carbs_mg"`
					CarbsSugarsMg  int `json:"carbs_sugars_mg"`
					FibreMg        int `json:"fibre_mg"`
					ProteinMg      int `json:"protein_mg"`
					SaltMg         int `json:"salt_mg"`
					NetWeightMg    int `json:"net_weight_mg"`
				} `json:"per_hundred_grams"`
				PerPortion struct {
					EnergyKcal     int `json:"energy_kcal"`
					EnergyKj       int `json:"energy_kj"`
					FatMg          int `json:"fat_mg"`
					FatSaturatesMg int `json:"fat_saturates_mg"`
					CarbsMg        int `json:"carbs_mg"`
					CarbsSugarsMg  int `json:"carbs_sugars_mg"`
					FibreMg        int `json:"fibre_mg"`
					ProteinMg      int `json:"protein_mg"`
					SaltMg         int `json:"salt_mg"`
					NetWeightMg    int `json:"net_weight_mg"`
				} `json:"per_portion"`
			} `json:"nutritional_information"`
		} `json:"entry"`
	} `json:"data"`
}

func ScrapeGousto(data *model.DataInput, r *model.Recipe) error {
	u, err := url.Parse(data.Url)
	if err != nil {
		return fmt.Errorf("error parsing url: %w", err)
	}

	parts := strings.Split(u.Path, "/")
	recipeId := parts[len(parts)-1]

	body, _, err := utils.ReadUrl("https://production-api.gousto.co.uk/cmsreadbroker/v1/recipe/"+recipeId, map[string][]string{
		"Accept":     {"application/json"},
		"User-Agent": {"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/110.0"},
	})
	if err != nil {
		return err
	}

	goustoData := GoustoData{}
	if err := json.Unmarshal(body, &goustoData); err != nil {
		return err
	}

	if err := parseGoustoData(&goustoData, r); err != nil {
		return err
	}

	return nil
}

func parseGoustoData(data *GoustoData, r *model.Recipe) error {
	if data.Status != "ok" {
		return errors.New("status is not ok")
	}

	if data.Data.Entry.Title != "" {
		r.Name = data.Data.Entry.Title
	}

	if data.Data.Entry.Description != "" {
		r.Description = data.Data.Entry.Description
	}

	if len(data.Data.Entry.Categories) != 0 {
		r.Categories = utils.AppendUnique(r.Categories, data.Data.Entry.Categories[0].Title)
	}

	if len(data.Data.Entry.Cuisine.Title) != 0 {
		r.Cuisines = utils.AppendUnique(r.Cuisines, data.Data.Entry.Cuisine.Title)
	}

	if len(data.Data.Entry.Media.Images) != 0 {
		for _, item := range data.Data.Entry.Media.Images {
			var image model.ImageObject
			image.Url = item.Image
			image.Width = item.Width
			r.AddImage(&image)
		}
	}

	if data.Data.Entry.Rating.Count > 0 {
		r.Rating = &model.AggregateRating{}
		r.Rating.RatingCount = data.Data.Entry.Rating.Count
		r.Rating.RatingValue = data.Data.Entry.Rating.Average
	}

	if data.Data.Entry.PrepTimes.For2 > 0 {
		r.CookTime = duration.Format(time.Duration(data.Data.Entry.PrepTimes.For2) * time.Minute)
	}

	if len(data.Data.Entry.Ingredients) != 0 {
		r.Yield = 2
		for _, item := range data.Data.Entry.Ingredients {
			r.Ingredients = utils.AppendUnique(r.Ingredients, item.Name)
		}
		for _, item := range data.Data.Entry.Basics {
			r.Ingredients = utils.AppendUnique(r.Ingredients, item.Title)
		}
	}

	if len(data.Data.Entry.CookingInstructions) != 0 {
		for _, item := range data.Data.Entry.CookingInstructions {
			var step model.HowToStep
			step.Text = utils.CleanupInline(item.Instruction)
			if len(item.Media.Images) != 0 {
				step.Image = utils.CleanupInline(item.Media.Images[0].Image)
			}
			r.Instructions = append(r.Instructions, &model.HowToSection{HowToStep: step})
		}
	}

	return nil
}
