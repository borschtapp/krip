package schema

import (
	"fmt"
	"net/url"
	"time"

	"github.com/sosodev/duration"

	"github.com/astappiev/microdata"
	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/utils"
)

func Scrape(data *model.DataInput, r *model.Recipe) error {
	if data.Schemas == nil {
		return nil
	}
	baseUrl, _ := url.Parse(r.Url)

	recipeSchema := data.Schemas.GetFirstOfSchemaType("Recipe")
	if recipeSchema != nil {
		parseRecipe(recipeSchema, r, baseUrl)
	}

	siteSchema := data.Schemas.GetFirstOfSchemaType("WebSite")
	if siteSchema != nil {
		parsePublisher(siteSchema, r.Publisher, baseUrl, false)
	}
	if recipeSchema != nil {
		if item, ok := recipeSchema.GetNestedItem("publisher", "brand"); ok {
			parsePublisher(item, r.Publisher, baseUrl, true)
		} else if val, ok := getPropertyString(recipeSchema, "publisher", "brand"); ok {
			r.Publisher.Name = utils.CleanupInline(val)
		}
	}
	orgSchema := data.Schemas.GetFirstOfSchemaType("Organization")
	if orgSchema != nil {
		parsePublisher(orgSchema, r.Publisher, baseUrl, false)
	}
	estSchema := data.Schemas.GetFirstOfSchemaType("FoodEstablishment")
	if estSchema != nil {
		parsePublisher(estSchema, r.Publisher, baseUrl, false)
	}

	if recipeSchema != nil {
		if item, ok := recipeSchema.GetNestedItem("author", "creator"); ok {
			parseAuthor(item, r.Author, baseUrl, true)
		} else if val, ok := getPropertyString(recipeSchema, "author", "creator"); ok {
			r.Author.Name = utils.CleanupInline(val)
		}
	}
	personSchema := data.Schemas.GetFirstOfSchemaType("Person")
	if personSchema != nil {
		parseAuthor(personSchema, r.Author, baseUrl, r.Publisher != nil && r.Author != nil && r.Publisher.Name == r.Author.Name)
	}

	return nil
}

func parseRecipe(recipeSchema *microdata.Item, r *model.Recipe, baseUrl *url.URL) {
	if val, ok := getPropertyString(recipeSchema, "url", "URL"); ok && r.Url != val && utils.IsAbsolute(val) {
		r.Url = val
	}

	if val, ok := getPropertyString(recipeSchema, "name", "headline"); ok {
		r.Name = utils.CleanupInline(val)
	}

	if values, ok := getPropertiesKeywords(recipeSchema, "recipeCategory"); ok {
		r.Categories = values
	}

	if val, ok := getPropertyDuration(recipeSchema, "totalTime", "TotalTime"); ok {
		r.TotalTime = duration.Format(val)
	}

	if val, ok := getPropertyDuration(recipeSchema, "cookTime", "CookTime", "performTime"); ok {
		r.CookTime = duration.Format(val)
	}

	if val, ok := getPropertyDuration(recipeSchema, "prepTime", "PrepTime"); ok {
		r.PrepTime = duration.Format(val)
	}

	if val, ok := recipeSchema.GetProperty("recipeYield", "yield"); ok {
		switch v := val.(type) {
		case string:
			r.Yield = int(utils.FindNumber(v))
		case float64:
			r.Yield = int(v)
		default:
			fmt.Println("unable to parse recipeYield: ", fmt.Sprint(v))
		}
	}

	if nested, ok := recipeSchema.GetNested("image"); ok {
		for _, item := range nested.Items {
			image := &model.ImageObject{}
			if val, ok := getPropertyString(item, "url"); ok {
				image.Url = utils.ToAbsoluteUrl(baseUrl, val)
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
	} else if values, ok := getPropertiesArray(recipeSchema, "image"); ok {
		for _, val := range values {
			r.AddImageUrl(utils.ToAbsoluteUrl(baseUrl, val))
		}
	} else if val, ok := getPropertyString(recipeSchema, "thumbnailUrl"); ok {
		r.AddImageUrl(utils.ToAbsoluteUrl(baseUrl, val))
	}

	if item, ok := recipeSchema.GetNestedItem("nutrition"); ok {
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

	if val, ok := getPropertyString(recipeSchema, "inLanguage", "language"); ok {
		r.Language = utils.CleanupLang(val)
	}

	if val, ok := getPropertyString(recipeSchema, "articleBody", "articleSection", "about"); ok {
		r.Text = utils.Cleanup(val)
	}

	if values, ok := recipeSchema.GetProperties("recipeIngredient", "ingredients", "supply"); ok {
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

	if values, ok := recipeSchema.GetProperties("tool"); ok {
		for _, val := range values {
			if val, ok := getStringOrChild(val, "name"); ok {
				r.Equipment = append(r.Equipment, val)
			}
		}
	}

	if nested, ok := recipeSchema.GetNested("recipeInstructions", "instructions", "step"); ok {
		for _, item := range nested.Items {
			if item.IsOfSchemaType("HowToStep") {
				// yummly stores publisher in every step, but not in root of the schema
				if val, ok := item.GetNestedItem("publisher"); ok {
					parsePublisher(val, r.Publisher, baseUrl, true)
				}
				if val, ok := item.GetNestedItem("author"); ok {
					parseAuthor(val, r.Author, baseUrl, true)
				}

				r.Instructions = append(r.Instructions, &model.HowToSection{HowToStep: parseHowToStep(item)})
			} else if item.IsOfSchemaType("HowToSection") {
				section := model.HowToSection{HowToStep: parseHowToStep(item)}
				if nested, ok := item.GetNested("itemListElement", "ItemListElement"); ok {
					for _, item := range nested.Items {
						var step = parseHowToStep(item)
						section.Steps = append(section.Steps, &step)
					}
				}
				r.Instructions = append(r.Instructions, &section)
			} else if item.IsOfSchemaType("ItemList") {
				if nested, ok := item.GetNested("itemListElement", "ItemListElement"); ok {
					for _, item := range nested.Items {
						r.Instructions = append(r.Instructions, &model.HowToSection{HowToStep: parseHowToStep(item)})
					}
				}
			} else {
				fmt.Println("unknown instruction type: ", fmt.Sprint(item.Types))
			}
		}
	} else if values, ok := getPropertiesArray(recipeSchema, "recipeInstructions", "instructions"); ok {
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

	if item, ok := recipeSchema.GetNestedItem("aggregateRating"); ok {
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

	if values, ok := getPropertiesKeywords(recipeSchema, "recipeCuisine"); ok {
		r.Cuisines = values
	}

	if val, ok := getPropertyString(recipeSchema, "cookingMethod", "CookingMethod"); ok {
		r.CookingMethod = utils.CleanupInline(val)
	}

	if val, ok := getPropertyInt(recipeSchema, "commentCount"); ok {
		r.CommentCount = val
	}

	if val, ok := getPropertyString(recipeSchema, "suitableForDiet"); ok {
		r.Diets = utils.AppendUnique(r.Diets, utils.CleanupInline(val))
	}

	if val, ok := getPropertyString(recipeSchema, "description"); ok {
		r.Description = utils.CleanupInline(val)
	}

	if values, ok := getPropertiesKeywords(recipeSchema, "keywords", "Keywords"); ok {
		r.Keywords = values
	}

	if item, ok := recipeSchema.GetNestedItem("video"); ok {
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
			video.EmbedUrl = utils.ToAbsoluteUrl(baseUrl, val)
		}
		if val, ok := getPropertyString(item, "contentURL", "contentUrl"); ok {
			video.ContentUrl = utils.ToAbsoluteUrl(baseUrl, val)
		}
		if val, ok := getPropertyString(item, "thumbnailUrl", "image"); ok {
			video.ThumbnailUrl = utils.ToAbsoluteUrl(baseUrl, val)
		}
		if val, ok := getPropertyString(item, "uploadDate", "datePublished"); ok {
			if val, err := time.Parse(time.RFC3339, val); err == nil {
				video.UploadDate = &val
			}
		}
		r.Video = video
	}

	if val, ok := getPropertyString(recipeSchema, "datePublished", "dateCreated"); ok {
		if val, err := time.Parse(time.RFC3339, val); err == nil {
			r.DatePublished = &val
		}
	}

	if val, ok := getPropertyString(recipeSchema, "dateModified"); ok {
		if val, err := time.Parse(time.RFC3339, val); err == nil {
			r.DateModified = &val
		}
	}
}

func parsePublisher(item *microdata.Item, o *model.Organization, baseUrl *url.URL, override bool) {
	if val, ok := getPropertyString(item, "name"); ok && (override || len(o.Name) == 0) {
		o.Name = utils.CleanupInline(val)
	}
	if val, ok := getPropertyString(item, "url"); ok && (override || len(o.Url) == 0) {
		o.Url = utils.RemoveTrailingSlash(val)
	}
	if val, ok := getPropertyString(item, "description"); ok && (override || len(o.Description) == 0) {
		o.Description = utils.CleanupInline(val)
	}
	if val, ok := getPropertyStringOrChild(item, "logo", "url"); ok && (override || len(o.Logo) == 0) {
		o.Logo = utils.ToAbsoluteUrl(baseUrl, val)
	}
}

func parseAuthor(item *microdata.Item, p *model.Person, baseUrl *url.URL, override bool) {
	if val, ok := getPropertyString(item, "name", "Name", "alternateName"); ok && (override || len(p.Name) == 0) {
		p.Name = utils.CleanupInline(val)
	}
	if val, ok := getPropertyString(item, "jobTitle", "JobTitle"); ok && (override || len(p.JobTitle) == 0) {
		p.JobTitle = utils.CleanupInline(val)
	}
	if val, ok := getPropertiesArray(item, "knowsAbout", "KnowsAbout"); ok && (override || len(p.KnowsAbout) == 0) {
		p.KnowsAbout = val
	} else if val, ok := getPropertyString(item, "knowsAbout"); ok {
		p.KnowsAbout = utils.AppendUnique(p.KnowsAbout, utils.CleanupInline(val))
	}
	if val, ok := getPropertyString(item, "description", "about"); ok && (override || len(p.Description) == 0) {
		p.Description = utils.CleanupInline(val)
	}
	if val, ok := getPropertyString(item, "url"); ok && (override || len(p.Url) == 0) {
		p.Url = utils.ToAbsoluteUrl(baseUrl, val)
	}
	if val, ok := getPropertyStringOrChild(item, "image", "url"); ok && (override || len(p.Image) == 0) {
		p.Image = utils.ToAbsoluteUrl(baseUrl, val)
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

func ScrapeFeed(data *model.DataInput, feed *model.Feed) error {
	if data.Schemas == nil {
		return fmt.Errorf("no schemas found")
	}

	baseUrl, _ := url.Parse(data.Url)

	// Extracting publisher is kind of easy here
	publisher := &model.Organization{}
	if siteSchema := data.Schemas.GetFirstOfSchemaType("WebSite"); siteSchema != nil {
		parsePublisher(siteSchema, publisher, baseUrl, false)
	}
	if orgSchema := data.Schemas.GetFirstOfSchemaType("Organization"); orgSchema != nil {
		parsePublisher(orgSchema, publisher, baseUrl, false)
	}
	if estSchema := data.Schemas.GetFirstOfSchemaType("FoodEstablishment"); estSchema != nil {
		parsePublisher(estSchema, publisher, baseUrl, false)
	}

	// Now, let's try to guess where recipes can be hidden
	for _, item := range data.Schemas.Items {
		// An ItemList with Recipes (ideal scenario, newer seen that yet)
		if item.IsOfSchemaType("ItemList") {
			if nested, ok := item.GetNested("itemListElement"); ok {
				for _, item := range nested.Items {
					if item.IsOfSchemaType("Recipe") {
						recipe := &model.Recipe{Publisher: publisher}
						parseRecipe(item, recipe, baseUrl)
						feed.AddEntry(recipe)
					}
				}
			}
		}

		// Many recipes on the page itself
		if item.IsOfSchemaType("Recipe") {
			// A recipe that consists of another recipes, case of https://foodnetwork.co.uk/collections/air-fryer-recipes
			if nested, ok := item.GetNested("hasPart"); ok {
				for _, item := range nested.Items {
					if item.IsOfSchemaType("Recipe") {
						recipe := &model.Recipe{Publisher: publisher}
						parseRecipe(item, recipe, baseUrl)
						feed.AddEntry(recipe)
					}
				}
			} else {
				recipe := &model.Recipe{Publisher: publisher}
				parseRecipe(item, recipe, baseUrl)
				feed.AddEntry(recipe)
			}
		}
	}

	return nil
}
