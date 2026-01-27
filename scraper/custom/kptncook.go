package custom

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/utils"
	"github.com/sosodev/duration"
)

const kptnKey = "6q7QNKy-oIgk-IMuWisJ-jfN7s6"

type KptnCookRecipe struct {
	Title           string `json:"title"`
	Rtype           string `json:"rtype"`
	Gdocs           string `json:"gdocs"`
	AuthorComment   string `json:"authorComment"`
	UID             string `json:"uid"`
	Country         string `json:"country"`
	OtherIngred     string `json:"otherIngred"`
	PreparationTime int    `json:"preparationTime"`
	CookingTime     int    `json:"cookingTime"`
	RecipeNutrition struct {
		Calories     float64 `json:"calories"`
		Protein      float64 `json:"protein"`
		Fat          float64 `json:"fat"`
		Carbohydrate float64 `json:"carbohydrate"`
	} `json:"recipeNutrition"`
	ActiveTags []string `json:"activeTags"`
	Steps      []struct {
		Title string `json:"title"`
		Image struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"image"`
		Ingredients []struct {
			IngredientID string `json:"ingredientId"`
			Title        string `json:"title"`
			NumberTitle  struct {
				Singular string `json:"singular"`
				Plural   string `json:"plural"`
			} `json:"numberTitle"`
			Unit struct {
				Quantity float64 `json:"quantity"`
				Measure  string  `json:"measure"`
			} `json:"unit,omitempty"`
		} `json:"ingredients,omitempty"`
	} `json:"steps"`
	Authors []struct {
		ID struct {
			Oid string `json:"$oid"`
		} `json:"_id"`
		Name        string `json:"name"`
		Link        string `json:"link"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Facebook    string `json:"facebook"`
		Instagram   string `json:"instagram"`
		Twitter     string `json:"twitter"`
		Sponsor     string `json:"sponsor"`
		AuthorImage struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"authorImage"`
		CreationDate struct {
			Date int64 `json:"$date"`
		} `json:"creationDate"`
		UpdateDate struct {
			Date int64 `json:"$date"`
		} `json:"updateDate"`
	} `json:"authors"`
	Ingredients []struct {
		Quantity                float64 `json:"quantity,omitempty"`
		Measure                 string  `json:"measure,omitempty"`
		MetricQuantity          float64 `json:"metricQuantity,omitempty"`
		MetricMeasure           string  `json:"metricMeasure,omitempty"`
		QuantityUS              float64 `json:"quantityUS,omitempty"`
		MeasureUS               string  `json:"measureUS,omitempty"`
		ImperialQuantity        float64 `json:"imperialQuantity,omitempty"`
		ImperialMeasure         string  `json:"imperialMeasure,omitempty"`
		QuantityUSProd          float64 `json:"quantityUSProd,omitempty"`
		MeasureUSProd           string  `json:"measureUSProd,omitempty"`
		ImperialProductQuantity float64 `json:"imperialProductQuantity,omitempty"`
		ImperialProductMeasure  string  `json:"imperialProductMeasure,omitempty"`
		Ingredient              struct {
			ID struct {
				Oid string `json:"$oid"`
			} `json:"_id"`
			Typ         string `json:"typ"`
			Title       string `json:"title"`
			NumberTitle struct {
				Singular string `json:"singular"`
				Plural   string `json:"plural"`
			} `json:"numberTitle"`
			UncountableTitle string `json:"uncountableTitle"`
			Category         string `json:"category"`
			Key              string `json:"key"`
			Desc             string `json:"desc"`
			Image            struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"image"`
			IsSponsored bool `json:"isSponsored"`
			Measures    struct {
				De []string `json:"de"`
				Us []string `json:"us"`
			} `json:"measures"`
			Synonym string `json:"synonym"`
			Brands  []struct {
				ID              string   `json:"id"`
				Name            string   `json:"name"`
				Countries       []string `json:"countries"`
				IngredientTitle struct {
					Singular string `json:"singular"`
					Plural   string `json:"plural"`
				} `json:"ingredientTitle"`
				UncountableTitle string `json:"uncountableTitle"`
				IngredientImage  struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				} `json:"ingredientImage"`
			} `json:"brands"`
			CreationDate struct {
				Date int64 `json:"$date"`
			} `json:"creationDate"`
			UpdateDate struct {
				Date int64 `json:"$date"`
			} `json:"updateDate"`
		} `json:"ingredient,omitempty"`
	} `json:"ingredients"`
	ImageList []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
		Type string `json:"type"`
	} `json:"imageList"`
	LocalizedPublishDate struct {
		En struct {
			Date int64 `json:"$date"`
		} `json:"en"`
		De struct {
			Date int64 `json:"$date"`
		} `json:"de"`
	} `json:"localizedPublishDate"`
	TrackingMode    string `json:"trackingMode"`
	Feature         string `json:"feature"`
	PublishDuration struct {
		En int `json:"en"`
		De int `json:"de"`
	} `json:"publishDuration"`
	IngredientTags string `json:"ingredientTags"`
	FavoriteCount  int    `json:"favoriteCount"`
	PublishDates   struct {
		En []struct {
			Date int64 `json:"$date"`
		} `json:"en"`
		De []struct {
			Date int64 `json:"$date"`
		} `json:"de"`
	} `json:"publishDates"`
	CreationDate struct {
		Date int64 `json:"$date"`
	} `json:"creationDate"`
	UpdateDate struct {
		Date int64 `json:"$date"`
	} `json:"updateDate"`
}

func ScrapeKptnCook(data *model.DataInput, r *model.Recipe) error {
	u, err := url.Parse(data.Url)
	if err != nil {
		return fmt.Errorf("error parsing url: %w", err)
	}

	if lang := u.Query().Get("lang"); lang != "" {
		r.Language = lang
	} else {
		r.Language = "en"
	}

	parts := strings.Split(u.Path, "/")
	recipeId := parts[len(parts)-1]

	jsonBody := []byte(`[{"uid": "` + recipeId + `"}]`)
	bodyReader := bytes.NewReader(jsonBody)

	client := http.Client{}
	req, err := http.NewRequest("POST", "https://mobile.kptncook.com/recipes/search?kptnkey="+kptnKey+"&lang="+r.Language, bodyReader)
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	req.Header = map[string][]string{
		"Content-Type": {"application/json"},
		"Accept":       {"application/json"},
		"User-Agent":   {"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/110.0"},
	}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("could not send request: %w", err)
	}

	if res.Body != nil {
		defer res.Body.Close()
		body, readErr := io.ReadAll(res.Body)
		if readErr != nil {
			return errors.New("could not read response body: " + readErr.Error())
		}

		if res.StatusCode != 200 {
			return errors.New("invalid status " + res.Status + ": " + string(body))
		}

		var kptnData []KptnCookRecipe
		if err := json.Unmarshal(body, &kptnData); err != nil {
			return err
		}

		if err := parseKptnData(&kptnData[0], r); err != nil {
			return err
		}
	}

	return nil
}

func parseKptnData(data *KptnCookRecipe, r *model.Recipe) error {
	if len(data.Title) != 0 {
		r.Name = utils.CleanupInline(data.Title)
	}

	if len(data.Rtype) != 0 {
		r.Categories = utils.AppendUnique(r.Categories, data.Rtype)
	}

	if data.PreparationTime != 0 {
		r.TotalTime = duration.Format(time.Duration(data.PreparationTime) * time.Minute)
	}

	if data.CookingTime != 0 {
		r.CookTime = duration.Format(time.Duration(data.CookingTime) * time.Minute)
	}

	r.Nutrition = &model.NutritionInformation{}
	if data.RecipeNutrition.Fat > 0 {
		r.Nutrition.FatContent = data.RecipeNutrition.Fat
	}
	if data.RecipeNutrition.Protein > 0 {
		r.Nutrition.ProteinContent = data.RecipeNutrition.Protein
	}
	if data.RecipeNutrition.Calories > 0 {
		r.Nutrition.Calories = data.RecipeNutrition.Calories
	}
	if data.RecipeNutrition.Carbohydrate > 0 {
		r.Nutrition.CarbohydrateContent = data.RecipeNutrition.Carbohydrate
	}

	if len(data.ActiveTags) != 0 {
		for _, tag := range data.ActiveTags {
			r.Keywords = utils.AppendUnique(r.Keywords, tag)
		}
	}

	if len(data.Steps) != 0 {
		r.Instructions = []*model.HowToSection{}
		for _, item := range data.Steps {
			var step model.HowToStep
			step.Text = utils.CleanupInline(item.Title)
			if item.Image.URL != "" {
				step.Image = item.Image.URL + "?kptnkey=" + kptnKey
			}
			r.Instructions = append(r.Instructions, &model.HowToSection{HowToStep: step})
		}
	}

	if len(data.Authors) != 0 {
		for _, item := range data.Authors {
			if item.Name != "" {
				r.Author.Name = item.Name
			}
			if item.Link != "" {
				r.Author.Url = item.Link
			}
			if item.Description != "" {
				r.Author.Description = item.Description
			}
			if item.Title != "" {
				r.Author.JobTitle = item.Title
			}
			if item.AuthorImage.URL != "" {
				r.Author.Image = item.AuthorImage.URL
			}
		}
	}

	if len(data.Ingredients) != 0 {
		r.Yield = 1
		r.Ingredients = []string{}
		for _, item := range data.Ingredients {
			if item.Quantity == 0 {
				r.Ingredients = append(r.Ingredients, item.Ingredient.UncountableTitle)
			} else {
				r.Ingredients = append(r.Ingredients, fmt.Sprintf("%v %s %s", item.MetricQuantity, item.MetricMeasure, item.Ingredient.UncountableTitle))
			}
		}
	}

	if len(data.ImageList) != 0 {
		for _, item := range data.ImageList {
			r.AddImageUrl(item.URL + "?kptnkey=" + kptnKey)
		}
	}

	if data.CreationDate.Date > 0 {
		pubDate := time.Unix(data.CreationDate.Date/1000, 0)
		r.DatePublished = &pubDate
	}

	if data.FavoriteCount > 0 {
		r.Rating = &model.AggregateRating{}
		r.Rating.RatingCount = data.FavoriteCount
	}

	return nil
}
