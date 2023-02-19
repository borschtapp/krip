package schema

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/astappiev/microdata"
	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/utils"
)

func Scrape(data *model.DataInput, r *model.Recipe) error {
	if data.Schema == nil {
		return nil
	}

	if val, ok := getPropertyString(data.Schema, "url", "URL"); ok && strings.HasPrefix(val, "http") {
		r.Url = val
	}

	if val, ok := getPropertyString(data.Schema, "name", "headline"); ok {
		r.Name = utils.CleanupInline(val)
	}

	if values, ok := getPropertiesKeywords(data.Schema, "recipeCategory"); ok {
		r.Categories = values
	}

	if val, ok := getPropertyDuration(data.Schema, "totalTime", "TotalTime"); ok {
		r.TotalTime = val.Minutes()
	}

	if val, ok := getPropertyDuration(data.Schema, "cookTime", "CookTime", "performTime"); ok {
		r.CookTime = val.Minutes()
	}

	if val, ok := getPropertyDuration(data.Schema, "prepTime", "PrepTime"); ok {
		r.PrepTime = val.Minutes()
	}

	if val, ok := data.Schema.GetProperty("recipeYield", "yield"); ok {
		switch val.(type) {
		case string:
			r.Yield = int(utils.FindNumber(val.(string)))
		case float64:
			r.Yield = int(val.(float64))
		default:
			return errors.New("unable to parse recipeYield: " + fmt.Sprint(val))
		}
	}

	if nested, ok := data.Schema.GetNested("image"); ok {
		for _, item := range nested.Items {
			image := &model.ImageObject{}
			if val, ok := getPropertyString(item, "url"); ok {
				image.Url = val
			}
			if val, ok := getPropertyInt(item, "width"); ok {
				image.Width = val
			}
			if val, ok := getPropertyInt(item, "height"); ok {
				image.Height = val
			}
			if val, ok := getPropertyString(item, "caption"); ok {
				image.Caption = utils.CleanupInline(val)
			}
			r.AddImage(image)
		}
	} else if values, ok := getPropertiesArray(data.Schema, "image"); ok {
		for _, val := range values {
			r.AddImage(&model.ImageObject{Url: val})
		}
	}

	if val, ok := getPropertyString(data.Schema, "thumbnailUrl"); ok {
		r.ThumbnailUrl = val
	}

	if item, ok := data.Schema.GetNestedItem("nutrition"); ok {
		r.Nutrition = &model.NutritionInformation{}
		for key, val := range item.Properties {
			strVal := fmt.Sprint(val[0])

			switch key {
			case "calories":
				r.Nutrition.Calories = strVal
			case "servingSize":
				r.Nutrition.ServingSize = strVal
			case "carbohydrateContent":
				r.Nutrition.CarbohydrateContent = strVal
			case "cholesterolContent":
				r.Nutrition.CholesterolContent = strVal
			case "fatContent":
				r.Nutrition.FatContent = strVal
			case "fiberContent":
				r.Nutrition.FiberContent = strVal
			case "proteinContent":
				r.Nutrition.ProteinContent = strVal
			case "saturatedFatContent":
				r.Nutrition.SaturatedFatContent = strVal
			case "sodiumContent":
				r.Nutrition.SodiumContent = strVal
			case "sugarContent":
				r.Nutrition.SugarContent = strVal
			case "transFatContent":
				r.Nutrition.TransFatContent = strVal
			case "unsaturatedFatContent":
				r.Nutrition.UnsaturatedFatContent = strVal
			}
		}
	}

	if val, ok := getPropertyString(data.Schema, "inLanguage", "language"); ok {
		r.Language = val
	}

	if val, ok := getPropertyString(data.Schema, "articleBody", "articleSection", "about"); ok {
		r.Text = utils.Cleanup(val)
	}

	if values, ok := data.Schema.GetProperties("recipeIngredient", "ingredients", "supply"); ok {
		for _, val := range values {
			if text, item := getStringOrItem(val); len(text) != 0 {
				r.Ingredients = append(r.Ingredients, text)
			} else if item != nil {
				if name, ok := getPropertyString(item, "name"); ok {
					name = utils.CleanupInline(name)
					if amount, ok := getPropertyString(item, "amount"); ok {
						name = utils.CleanupInline(amount) + " " + name
					}
					r.Ingredients = append(r.Ingredients, name)
				}
			}
		}
	}

	if values, ok := data.Schema.GetProperties("tool"); ok {
		for _, val := range values {
			if val, ok := getStringOrChild(val, "name"); ok {
				r.Equipment = append(r.Equipment, val)
			}
		}
	}

	if nested, ok := data.Schema.GetNested("recipeInstructions", "instructions", "step"); ok {
		for _, item := range nested.Items {
			if item.IsOfType("HowToStep", "http://schema.org/HowToStep", "https://schema.org/HowToStep") {
				// yummly stores publisher in every step, but not in root of the schema
				if val, ok := item.GetNestedItem("publisher"); ok {
					parsePublisher(val, r, false)
				}

				r.Instructions = append(r.Instructions, &model.HowToSection{HowToStep: parseHowToStep(item)})
			} else if item.IsOfType("HowToSection", "http://schema.org/HowToSection", "https://schema.org/HowToSection") {
				section := model.HowToSection{HowToStep: parseHowToStep(item)}
				if nested, ok := item.GetNested("itemListElement", "ItemListElement"); ok {
					for _, item := range nested.Items {
						var step = parseHowToStep(item)
						section.Steps = append(section.Steps, &step)
					}
				}
				r.Instructions = append(r.Instructions, &section)
			} else if item.IsOfType("ItemList", "http://schema.org/ItemList", "https://schema.org/ItemList") {
				if nested, ok := item.GetNested("itemListElement", "ItemListElement"); ok {
					for _, item := range nested.Items {
						r.Instructions = append(r.Instructions, &model.HowToSection{HowToStep: parseHowToStep(item)})
					}
				}
			} else {
				return errors.New("unknown instruction type: " + fmt.Sprint(item.Types))
			}
		}
	} else if values, ok := getPropertiesArray(data.Schema, "recipeInstructions", "instructions"); ok {
		if len(values) == 1 {
			values = utils.SplitParagraphs(values[0])
		} else {
			for i, val := range values {
				values[i] = utils.CleanupInline(val)
			}
		}

		for _, step := range values {
			r.Instructions = append(r.Instructions, &model.HowToSection{HowToStep: model.HowToStep{Text: step}})
		}
	}

	if item, ok := data.Schema.GetNestedItem("aggregateRating"); ok {
		r.Rating = &model.AggregateRating{}
		if val, ok := getPropertyInt(item, "ratingCount"); ok {
			r.Rating.RatingCount = val
		}
		if val, ok := getPropertyFloat(item, "ratingValue"); ok {
			r.Rating.RatingValue = val
		}
		if val, ok := getPropertyInt(item, "bestRating"); ok {
			r.Rating.BestRating = val
		}
		if val, ok := getPropertyInt(item, "worstRating"); ok {
			r.Rating.WorstRating = val
		}
		if val, ok := getPropertyInt(item, "reviewCount"); ok {
			r.Rating.ReviewCount = val
		}
	}

	if values, ok := getPropertiesKeywords(data.Schema, "recipeCuisine"); ok {
		r.Cuisines = values
	}

	if val, ok := getPropertyString(data.Schema, "cookingMethod", "CookingMethod"); ok {
		r.CookingMethod = utils.CleanupInline(val)
	}

	if val, ok := getPropertyInt(data.Schema, "commentCount"); ok {
		r.CommentCount = val
	}

	if val, ok := getPropertyString(data.Schema, "suitableForDiet"); ok {
		r.Diets = utils.AppendUnique(r.Diets, utils.CleanupInline(val))
	}

	if val, ok := getPropertyString(data.Schema, "description"); ok {
		r.Description = utils.CleanupInline(val)
	}

	if values, ok := getPropertiesKeywords(data.Schema, "keywords", "Keywords"); ok {
		r.Keywords = values
	}

	if item, ok := data.Schema.GetNestedItem("publisher", "brand"); ok {
		parsePublisher(item, r, true)
	}

	if item, ok := data.Schema.GetNestedItem("author", "creator"); ok {
		person := &model.Person{}
		if val, ok := getPropertyString(item, "name", "Name", "alternateName"); ok {
			person.Name = utils.CleanupInline(val)
		}
		if val, ok := getPropertyString(item, "jobTitle", "JobTitle"); ok {
			person.JobTitle = utils.CleanupInline(val)
		}
		if val, ok := getPropertyString(item, "description", "about"); ok {
			person.Description = utils.CleanupInline(val)
		}
		if val, ok := getPropertyString(item, "url"); ok {
			person.Url = val
		}
		if val, ok := getPropertyStringOrChild(item, "image", "url"); ok {
			person.Image = val
		}
		if r.Publisher == nil || person.Name != r.Publisher.Name {
			r.Author = person
		}
	} else if val, ok := getPropertyString(data.Schema, "author", "creator"); ok {
		person := &model.Person{Name: utils.CleanupInline(val)}
		if r.Publisher == nil || person.Name != r.Publisher.Name {
			r.Author = person
		}
	}

	if item, ok := data.Schema.GetNestedItem("video"); ok {
		var video = &model.VideoObject{}
		if val, ok := getPropertyString(item, "name"); ok {
			video.Name = utils.CleanupInline(val)
		}
		if val, ok := getPropertyString(item, "description"); ok {
			video.Description = utils.CleanupInline(val)
		}
		if val, ok := getPropertyString(item, "duration"); ok {
			video.Duration = utils.CleanupInline(val)
		}
		if val, ok := getPropertyString(item, "embedUrl", "embedURL", "url"); ok {
			video.EmbedUrl = utils.CleanupInline(val)
		}
		if val, ok := getPropertyString(item, "contentURL", "contentUrl"); ok {
			video.ContentUrl = utils.CleanupInline(val)
		}
		if val, ok := getPropertyString(item, "thumbnailUrl", "image"); ok {
			video.ThumbnailUrl = utils.CleanupInline(val)
		}
		if val, ok := getPropertyString(item, "uploadDate", "datePublished"); ok {
			if val, err := time.Parse(time.RFC3339, val); err == nil {
				video.UploadDate = &val
			}
		}
		r.Video = video
	}

	if val, ok := getPropertyString(data.Schema, "datePublished", "dateCreated"); ok {
		if val, err := time.Parse(time.RFC3339, val); err == nil {
			r.DatePublished = &val
		}
	}

	if val, ok := getPropertyString(data.Schema, "dateModified"); ok {
		if val, err := time.Parse(time.RFC3339, val); err == nil {
			r.DateModified = &val
		}
	}

	return nil
}

func parsePublisher(item *microdata.Item, r *model.Recipe, merge bool) {
	if r.Publisher == nil {
		r.Publisher = &model.Organization{}
	} else if !merge {
		return
	}

	if val, ok := getPropertyString(item, "name"); ok {
		r.Publisher.Name = utils.CleanupInline(val)
	}
	if val, ok := getPropertyString(item, "url"); ok {
		r.Publisher.Url = val
	}
	if val, ok := getPropertyStringOrChild(item, "logo", "url"); ok {
		r.Publisher.Logo = val
	}
}

func parseHowToStep(item *microdata.Item) model.HowToStep {
	var step model.HowToStep
	if val, ok := getPropertyStringOrChild(item, "text", "result"); ok {
		step.Text = utils.Cleanup(val)
	} else if val, ok := getPropertyString(item, "description"); ok {
		step.Text = utils.Cleanup(val)
	}
	if val, ok := getPropertyString(item, "name", "Name"); ok {
		val = utils.CleanupInline(val)
		if val != step.Text {
			step.Name = val
		}
	}
	if val, ok := getPropertyStringOrChild(item, "image", "url"); ok {
		step.Image = val
	}
	if val, ok := getPropertyStringOrChild(item, "video", "embedUrl", "embedURL", "url"); ok {
		step.Video = val
	}
	if val, ok := getPropertyString(item, "url"); ok {
		step.Url = val
	}

	return step
}
