package website

import (
	"encoding/json"
	"github.com/sosodev/duration"
	"log"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/utils"
)

type KitchenStoriesRecipe struct {
	ID         string `json:"id,omitempty"`
	Title      string `json:"title,omitempty"`
	Slug       string `json:"slug,omitempty"`
	Type       string `json:"type,omitempty"`
	ContentID  string `json:"content_id,omitempty"`
	Difficulty string `json:"difficulty,omitempty"`
	Duration   struct {
		Preparation int `json:"preparation,omitempty"`
		Baking      int `json:"baking,omitempty"`
		Resting     int `json:"resting,omitempty"`
	} `json:"duration,omitempty"`
	Image struct {
		ID     string `json:"id,omitempty"`
		Width  int    `json:"width,omitempty"`
		Height int    `json:"height,omitempty"`
		URL    string `json:"url,omitempty"`
	} `json:"image,omitempty"`
	Author struct {
		ID          string `json:"id,omitempty"`
		Name        string `json:"name,omitempty"`
		Type        string `json:"type,omitempty"`
		NewType     string `json:"new_type,omitempty"`
		Slug        string `json:"slug,omitempty"`
		Occupation  string `json:"occupation,omitempty"`
		Description string `json:"description,omitempty"`
		Image       struct {
			URL string `json:"url,omitempty"`
		} `json:"image,omitempty"`
		Website     string `json:"website,omitempty"`
		BannerImage struct {
			URL string `json:"url,omitempty"`
		} `json:"banner_image,omitempty"`
		IsPremium bool `json:"is_premium,omitempty"`
	} `json:"author,omitempty"`
	Publishing struct {
		Created   string `json:"created,omitempty"`
		Updated   string `json:"updated,omitempty"`
		Published string `json:"published,omitempty"`
		State     string `json:"state,omitempty"`
	} `json:"publishing,omitempty"`
	URL           string `json:"url,omitempty"`
	UserReactions struct {
		Rating        float64 `json:"rating,omitempty"`
		RatingCount   int     `json:"rating_count,omitempty"`
		ImagesCount   int     `json:"images_count,omitempty"`
		CommentsCount int     `json:"comments_count,omitempty"`
		LikeCount     int     `json:"like_count,omitempty"`
		Quality       float64 `json:"quality,omitempty"`
	} `json:"user_reactions,omitempty"`
	Servings struct {
		Amount int    `json:"amount,omitempty"`
		Type   string `json:"type,omitempty"`
	} `json:"servings,omitempty"`
	ChefsNote string `json:"chefs_note,omitempty"`
	Nutrition struct {
		Calories     float64 `json:"calories,omitempty"`
		Fat          float64 `json:"fat,omitempty"`
		Protein      float64 `json:"protein,omitempty"`
		Carbohydrate float64 `json:"carbohydrate,omitempty"`
	} `json:"nutrition,omitempty"`
	Meta struct {
		Title       string `json:"title,omitempty"`
		Description string `json:"description,omitempty"`
		Hreflang    struct {
			En string `json:"en,omitempty"`
			De string `json:"de,omitempty"`
		} `json:"hreflang,omitempty"`
	} `json:"meta,omitempty"`
	Tags []struct {
		ID       string `json:"id,omitempty"`
		Slug     string `json:"slug,omitempty"`
		Title    string `json:"title,omitempty"`
		Type     string `json:"type,omitempty"`
		IsHidden bool   `json:"is_hidden,omitempty"`
	} `json:"tags,omitempty"`
	Categories struct {
		Main struct {
			ID    string `json:"id,omitempty"`
			Title string `json:"title,omitempty"`
			Slug  string `json:"slug,omitempty"`
			Path  []struct {
				ID    string `json:"id,omitempty"`
				Title string `json:"title,omitempty"`
				Slug  string `json:"slug,omitempty"`
			} `json:"path,omitempty"`
		} `json:"main,omitempty"`
		Additional []struct {
			ID    string `json:"id,omitempty"`
			Title string `json:"title,omitempty"`
			Slug  string `json:"slug,omitempty"`
			Path  []struct {
				ID    string `json:"id,omitempty"`
				Title string `json:"title,omitempty"`
				Slug  string `json:"slug,omitempty"`
			} `json:"path,omitempty"`
		} `json:"additional,omitempty"`
	} `json:"categories,omitempty"`
	HowtoVideos []struct {
		ID        string `json:"id,omitempty"`
		Title     string `json:"title,omitempty"`
		Slug      string `json:"slug,omitempty"`
		Type      string `json:"type,omitempty"`
		ContentID string `json:"content_id,omitempty"`
		RemoteID  string `json:"remote_id,omitempty"`
		URL       string `json:"url,omitempty"`
		Width     int    `json:"width,omitempty"`
		Height    int    `json:"height,omitempty"`
		Duration  int    `json:"duration,omitempty"`
		Image     struct {
			ID     string `json:"id,omitempty"`
			Width  int    `json:"width,omitempty"`
			Height int    `json:"height,omitempty"`
			URL    string `json:"url,omitempty"`
		} `json:"image,omitempty"`
		Publishing struct {
			Created       string `json:"created,omitempty"`
			Updated       string `json:"updated,omitempty"`
			Published     string `json:"published,omitempty"`
			State         string `json:"state,omitempty"`
			PlaybackState string `json:"playback-state,omitempty"`
		} `json:"publishing,omitempty"`
		Meta struct {
			Hreflang struct {
				En string `json:"en,omitempty"`
				De string `json:"de,omitempty"`
				Zh string `json:"zh,omitempty"`
			} `json:"hreflang,omitempty"`
		} `json:"meta,omitempty"`
		UserReactions struct {
			ViewCount int `json:"view_count,omitempty"`
		} `json:"user_reactions,omitempty"`
	} `json:"howto_videos,omitempty"`
	Ingredients []struct {
		List []struct {
			ID   string `json:"id,omitempty"`
			Name struct {
				Rendered string `json:"rendered,omitempty"`
				One      string `json:"one,omitempty"`
				Many     string `json:"many,omitempty"`
			} `json:"name,omitempty"`
			Measurement struct {
				Imperial struct {
					Amount float64 `json:"amount,omitempty"`
					Unit   struct {
						ID   string `json:"id,omitempty"`
						Name struct {
							One      string `json:"one,omitempty"`
							Many     string `json:"many,omitempty"`
							Rendered string `json:"rendered,omitempty"`
						} `json:"name,omitempty"`
						Type                   string `json:"type,omitempty"`
						IngredientPluralizable bool   `json:"ingredient_pluralizable,omitempty"`
					} `json:"unit,omitempty"`
				} `json:"imperial,omitempty"`
				Metric struct {
					Amount int `json:"amount,omitempty"`
					Unit   struct {
						ID   string `json:"id,omitempty"`
						Name struct {
							One      string `json:"one,omitempty"`
							Many     string `json:"many,omitempty"`
							Rendered string `json:"rendered,omitempty"`
						} `json:"name,omitempty"`
						Type                   string `json:"type,omitempty"`
						FeaturedOrder          int    `json:"featured_order,omitempty"`
						IngredientPluralizable bool   `json:"ingredient_pluralizable,omitempty"`
					} `json:"unit,omitempty"`
				} `json:"metric,omitempty"`
			} `json:"measurement,omitempty"`
			IsDivided bool `json:"is_divided,omitempty"`
			IsPartner bool `json:"is_partner,omitempty"`
		} `json:"list,omitempty"`
	} `json:"ingredients,omitempty"`
	Steps []struct {
		Text  string `json:"text,omitempty"`
		Image struct {
			URL string `json:"url,omitempty"`
		} `json:"image,omitempty"`
		Ingredients []struct {
			ID   string `json:"id,omitempty"`
			Name struct {
				Rendered string `json:"rendered,omitempty"`
				One      string `json:"one,omitempty"`
				Many     string `json:"many,omitempty"`
			} `json:"name,omitempty"`
			Measurement struct {
				Imperial struct {
					Amount float64 `json:"amount,omitempty"`
					Unit   struct {
						ID   string `json:"id,omitempty"`
						Name struct {
							One      string `json:"one,omitempty"`
							Many     string `json:"many,omitempty"`
							Rendered string `json:"rendered,omitempty"`
						} `json:"name,omitempty"`
						Type                   string `json:"type,omitempty"`
						IngredientPluralizable bool   `json:"ingredient_pluralizable,omitempty"`
					} `json:"unit,omitempty"`
				} `json:"imperial,omitempty"`
				Metric struct {
					Amount int `json:"amount,omitempty"`
					Unit   struct {
						ID   string `json:"id,omitempty"`
						Name struct {
							One      string `json:"one,omitempty"`
							Many     string `json:"many,omitempty"`
							Rendered string `json:"rendered,omitempty"`
						} `json:"name,omitempty"`
						Type                   string `json:"type,omitempty"`
						FeaturedOrder          int    `json:"featured_order,omitempty"`
						IngredientPluralizable bool   `json:"ingredient_pluralizable,omitempty"`
					} `json:"unit,omitempty"`
				} `json:"metric,omitempty"`
			} `json:"measurement,omitempty"`
			IsDivided bool `json:"is_divided,omitempty"`
			IsPartner bool `json:"is_partner,omitempty"`
		} `json:"ingredients,omitempty"`
		Utensils []struct {
			ID   string `json:"id,omitempty"`
			Name struct {
				Rendered string `json:"rendered,omitempty"`
				One      string `json:"one,omitempty"`
				Many     string `json:"many,omitempty"`
			} `json:"name,omitempty"`
			Size struct {
				ID   string `json:"id,omitempty"`
				Name string `json:"name,omitempty"`
			} `json:"size,omitempty"`
		} `json:"utensils,omitempty"`
	} `json:"steps,omitempty"`
	Utensils []struct {
		ID   string `json:"id,omitempty"`
		Name struct {
			Rendered string `json:"rendered,omitempty"`
			One      string `json:"one,omitempty"`
			Many     string `json:"many,omitempty"`
		} `json:"name,omitempty"`
		Size struct {
			ID   string `json:"id,omitempty"`
			Name string `json:"name,omitempty"`
		} `json:"size,omitempty"`
		Amount int `json:"amount,omitempty"`
	} `json:"utensils,omitempty"`
}

type KitchenStoriesScript struct {
	Props struct {
		PageProps struct {
			DehydratedState struct {
				Queries []struct {
					State struct {
						Data KitchenStoriesRecipe `json:"data,omitempty"`
					} `json:"state,omitempty"`
				} `json:"queries,omitempty"`
			} `json:"dehydratedState,omitempty"`
		} `json:"pageProps,omitempty"`
	} `json:"props,omitempty"`
}

func ScrapeKitchenStories(data *model.DataInput, r *model.Recipe) error {
	var ksRecipe *KitchenStoriesRecipe
	if data.Document != nil {
		data.Document.Find("script#__NEXT_DATA__").Each(func(i int, s *goquery.Selection) {
			var scriptData = &KitchenStoriesScript{}
			if err := json.Unmarshal([]byte(s.Text()), scriptData); err == nil {
				if len(scriptData.Props.PageProps.DehydratedState.Queries) > 0 {
					ksRecipe = &scriptData.Props.PageProps.DehydratedState.Queries[0].State.Data
				}
			} else {
				log.Println("error parsing kitchen stories json:", err)
			}
		})
	}

	if ksRecipe != nil {
		if len(ksRecipe.Title) != 0 {
			r.Name = ksRecipe.Title
		}
		if len(ksRecipe.Difficulty) != 0 {
			r.Difficulty = ksRecipe.Difficulty
		}
		if ksRecipe.Duration.Preparation != 0 {
			r.PrepTime = duration.Format(time.Duration(ksRecipe.Duration.Preparation) * time.Minute)
		}
		if ksRecipe.Duration.Baking != 0 {
			r.CookTime = duration.Format(time.Duration(ksRecipe.Duration.Baking) * time.Minute)
		}
		if len(ksRecipe.Image.URL) != 0 {
			r.AddImage(&model.ImageObject{Url: ksRecipe.Image.URL, Height: ksRecipe.Image.Height, Width: ksRecipe.Image.Width})
		}
		if len(ksRecipe.Author.Name) != 0 {
			r.Author = &model.Person{
				Name:        ksRecipe.Author.Name,
				Description: ksRecipe.Author.Description,
				JobTitle:    ksRecipe.Author.Occupation,
				Image:       ksRecipe.Author.Image.URL,
				Url:         ksRecipe.Author.Website,
			}
		}
		if len(ksRecipe.URL) != 0 {
			r.Url = ksRecipe.URL
		}
		if ksRecipe.UserReactions.CommentsCount != 0 {
			r.CommentCount = ksRecipe.UserReactions.CommentsCount
		}
		if ksRecipe.Servings.Amount != 0 {
			r.Yield = ksRecipe.Servings.Amount
		}
		if len(ksRecipe.ChefsNote) != 0 {
			r.Text = ksRecipe.ChefsNote
		}
		if len(ksRecipe.Meta.Description) != 0 {
			r.Description = ksRecipe.Meta.Description
		}
		if len(ksRecipe.Tags) != 0 {
			r.Keywords = nil
			for _, tag := range ksRecipe.Tags {
				if tag.Type == "diet" {
					r.Diets = utils.AppendUnique(r.Diets, utils.CleanupInline(tag.Title))
				} else if tag.Type == "ingredient" || tag.Type == "partner" || tag.Type == "recipe" || tag.Type == "monthly-issues" {
					continue // skip these tags
				} else {
					r.Keywords = utils.AppendUnique(r.Keywords, utils.CleanupInline(tag.Title))
				}
			}
		}
		if len(ksRecipe.Categories.Main.Title) != 0 {
			r.Categories = utils.AppendUnique(r.Categories, ksRecipe.Categories.Main.Title)
		}
		if len(ksRecipe.Categories.Additional) != 0 {
			for _, cat := range ksRecipe.Categories.Additional {
				r.Categories = utils.AppendUnique(r.Categories, cat.Title)
			}
		}
		if len(ksRecipe.Steps) != 0 {
			r.Instructions = nil

			for _, step := range ksRecipe.Steps {
				r.Instructions = append(r.Instructions, &model.HowToSection{HowToStep: model.HowToStep{
					Text:  step.Text,
					Image: step.Image.URL,
				}})
			}
		}
		if len(ksRecipe.Utensils) != 0 {
			for _, utensil := range ksRecipe.Utensils {
				r.Equipment = append(r.Equipment, utensil.Name.Rendered)
			}
		}
	}

	if r.Publisher == nil {
		r.Publisher = &model.Organization{}
		r.Publisher.Name = "Kitchen Stories"
		r.Publisher.Url = "https://www.kitchenstories.com"
	}

	if len(r.Publisher.Logo) == 0 {
		r.Publisher.Logo = "https://www.kitchenstories.com/images/ks-logo-2019.52cb693902fc25b0b22c9ee503355ec9.svg"
	}

	return nil
}
