package schema

import (
	"fmt"
	"log"
	"net/url"

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
	if val, ok := getPropertyString(recipeSchema, "description"); ok {
		r.Description = utils.CleanupInline(val)
	}
	if val, ok := getPropertyString(recipeSchema, "inLanguage", "language"); ok {
		r.Language = utils.CleanupLang(val)
	}
	if val, ok := getPropertyString(recipeSchema, "articleBody", "articleSection", "about"); ok {
		r.Text = utils.Cleanup(val)
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
			r.Yield = utils.CleanupInline(v)
		case float64:
			r.Yield = fmt.Sprint(v)
		default:
			log.Println("unable to parse recipeYield: ", fmt.Sprint(v))
		}
	}

	if values, ok := getPropertiesKeywords(recipeSchema, "recipeCategory"); ok {
		r.Categories = values
	}
	if values, ok := getPropertiesKeywords(recipeSchema, "recipeCuisine"); ok {
		r.Cuisines = values
	}
	if values, ok := getPropertiesKeywords(recipeSchema, "keywords", "Keywords"); ok {
		r.Keywords = values
	}
	if values, ok := getPropertiesArray(recipeSchema, "sameAs"); ok {
		r.Links = values
	}

	if val, ok := getPropertyString(recipeSchema, "cookingMethod", "CookingMethod"); ok {
		r.CookingMethod = utils.CleanupInline(val)
	}
	if val, ok := getPropertyString(recipeSchema, "educationalLevel", "difficulty"); ok {
		r.Difficulty = utils.CleanupInline(val)
	}
	if val, ok := getPropertyString(recipeSchema, "estimatedCost"); ok {
		r.EstimatedCost = utils.CleanupInline(val)
	}
	if val, ok := getPropertyString(recipeSchema, "suitableForDiet"); ok {
		r.Diets = utils.AppendUnique(r.Diets, utils.CleanupInline(val))
	}
	if val, ok := getPropertyInt(recipeSchema, "commentCount"); ok {
		r.CommentCount = val
	}

	if val, ok := getPropertyString(recipeSchema, "datePublished", "dateCreated"); ok {
		r.DatePublished, _ = utils.ParseRFC3339(val)
	}
	if val, ok := getPropertyString(recipeSchema, "dateModified"); ok {
		r.DateModified, _ = utils.ParseRFC3339(val)
	}

	parseRecipeImages(recipeSchema, r, baseUrl)
	parseRecipeNutrition(recipeSchema, r)
	parseRecipeIngredients(recipeSchema, r)
	parseRecipeEquipment(recipeSchema, r)
	parseRecipeInstructions(recipeSchema, r, baseUrl)
	parseRecipeRating(recipeSchema, r)
	parseRecipeVideo(recipeSchema, r, baseUrl)
}

func parseRecipeImages(item *microdata.Item, r *model.Recipe, baseUrl *url.URL) {
	if nested, ok := item.GetNested("image"); ok {
		for _, img := range nested.Items {
			image := &model.ImageObject{}
			if val, ok := getPropertyString(img, "url"); ok {
				image.Url = utils.ToAbsoluteUrl(baseUrl, val)
			}
			if val, ok := getPropertyInt(img, "width"); ok {
				image.Width = val
			}
			if val, ok := getPropertyInt(img, "height"); ok {
				image.Height = val
			}
			if val, ok := getPropertyString(img, "caption"); ok {
				image.Caption = utils.CleanupInline(val)
			}
			r.AddImage(image)
		}
	} else if values, ok := getPropertiesArray(item, "image"); ok {
		for _, val := range values {
			r.AddImageUrl(utils.ToAbsoluteUrl(baseUrl, val))
		}
	} else if val, ok := getPropertyString(item, "thumbnailUrl"); ok {
		r.AddImageUrl(utils.ToAbsoluteUrl(baseUrl, val))
	}
}

func parseRecipeNutrition(item *microdata.Item, r *model.Recipe) {
	nutrition, ok := item.GetNestedItem("nutrition")
	if !ok {
		return
	}

	r.Nutrition = &model.NutritionInformation{}
	for key, val := range nutrition.Properties {
		strVal := fmt.Sprint(val[0])
		switch key {
		case "calories":
			r.Nutrition.Calories = utils.FindNumber(strVal)
		case "servingSize":
			r.Nutrition.ServingSize = strVal
		case "carbohydrateContent":
			r.Nutrition.CarbohydrateContent = utils.FindNumber(strVal)
		case "cholesterolContent":
			r.Nutrition.CholesterolContent = utils.FindNumber(strVal)
		case "fatContent":
			r.Nutrition.FatContent = utils.FindNumber(strVal)
		case "fiberContent":
			r.Nutrition.FiberContent = utils.FindNumber(strVal)
		case "proteinContent":
			r.Nutrition.ProteinContent = utils.FindNumber(strVal)
		case "saturatedFatContent":
			r.Nutrition.SaturatedFatContent = utils.FindNumber(strVal)
		case "sodiumContent":
			r.Nutrition.SodiumContent = utils.FindNumber(strVal)
		case "sugarContent":
			r.Nutrition.SugarContent = utils.FindNumber(strVal)
		case "transFatContent":
			r.Nutrition.TransFatContent = utils.FindNumber(strVal)
		case "unsaturatedFatContent":
			r.Nutrition.UnsaturatedFatContent = utils.FindNumber(strVal)
		}
	}
}

func parseRecipeIngredients(item *microdata.Item, r *model.Recipe) {
	values, ok := item.GetProperties("recipeIngredient", "ingredients", "supply")
	if !ok {
		return
	}

	for _, val := range values {
		if text, ingredient := getStringOrItem(val); len(text) != 0 {
			r.Ingredients = append(r.Ingredients, &model.PropertyValue{Name: text})
		} else if ingredient != nil {
			prop := &model.PropertyValue{}
			if val, ok := getPropertyString(ingredient, "name", "item"); ok {
				prop.Name = utils.CleanupInline(val)
			}
			if val, ok := getPropertyString(ingredient, "amount", "value", "requiredQuantity"); ok {
				prop.Value = utils.CleanupInline(val)
			}
			if val, ok := getPropertyString(ingredient, "minValue", "MinValue"); ok {
				prop.MinValue = utils.CleanupInline(val)
			}
			if val, ok := getPropertyString(ingredient, "maxValue", "MaxValue"); ok {
				prop.MaxValue = utils.CleanupInline(val)
			}
			if val, ok := getPropertyString(ingredient, "unitCode", "UnitCode"); ok {
				prop.UnitCode = utils.CleanupInline(val)
			}
			if val, ok := getPropertyString(ingredient, "unitText", "UnitText"); ok {
				prop.UnitText = utils.CleanupInline(val)
			}
			if val, ok := getPropertyString(ingredient, "image", "Image"); ok {
				prop.Image = utils.CleanupInline(val)
			}
			if val, ok := getPropertyString(ingredient, "url", "Url"); ok {
				prop.Url = utils.CleanupInline(val)
			}
			if val, ok := getPropertyString(ingredient, "estimatedCost", "EstimatedCost"); ok {
				prop.EstimatedCost = utils.CleanupInline(val)
			}
			r.Ingredients = append(r.Ingredients, prop)
		}
	}
}

func parseRecipeEquipment(item *microdata.Item, r *model.Recipe) {
	values, ok := item.GetProperties("tool", "recipeEquipment")
	if !ok {
		return
	}

	for _, val := range values {
		if text, equipment := getStringOrItem(val); len(text) != 0 {
			r.Equipment = append(r.Equipment, &model.HowToTool{Name: text})
		} else if equipment != nil {
			tool := &model.HowToTool{}
			if val, ok := getPropertyString(equipment, "name", "item"); ok {
				tool.Name = val
			}
			if val, ok := getPropertyString(equipment, "description", "Description"); ok {
				tool.Description = val
			}
			if val, ok := getPropertyString(equipment, "url", "Url"); ok {
				tool.Url = val
			}
			if val, ok := getPropertyString(equipment, "image", "Image"); ok {
				tool.Image = val
			}
			if val, ok := getPropertyString(equipment, "requiredQuantity", "amount", "value"); ok {
				tool.Quantity = val
			}
			r.Equipment = append(r.Equipment, tool)
		}
	}
}

func parseRecipeInstructions(item *microdata.Item, r *model.Recipe, baseUrl *url.URL) {
	if nested, ok := item.GetNested("recipeInstructions", "instructions", "step"); ok {
		for _, step := range nested.Items {
			if step.IsOfSchemaType("HowToStep") {
				// yummly stores publisher in every step, but not in root of the schema
				if val, ok := step.GetNestedItem("publisher"); ok {
					parsePublisher(val, r.Publisher, baseUrl, true)
				}
				if val, ok := step.GetNestedItem("author"); ok {
					parseAuthor(val, r.Author, baseUrl, true)
				}
				r.Instructions = append(r.Instructions, &model.HowToSection{HowToStep: parseHowToStep(step)})
			} else if step.IsOfSchemaType("HowToSection") {
				section := model.HowToSection{HowToStep: parseHowToStep(step)}
				if nested, ok := step.GetNested("itemListElement", "ItemListElement"); ok {
					for _, s := range nested.Items {
						parsed := parseHowToStep(s)
						section.Steps = append(section.Steps, &parsed)
					}
				}
				r.Instructions = append(r.Instructions, &section)
			} else if step.IsOfSchemaType("ItemList") {
				if nested, ok := step.GetNested("itemListElement", "ItemListElement"); ok {
					for _, s := range nested.Items {
						r.Instructions = append(r.Instructions, &model.HowToSection{HowToStep: parseHowToStep(s)})
					}
				}
			} else {
				log.Println("unknown instruction type: ", fmt.Sprint(step.Types))
			}
		}
	} else if values, ok := getPropertiesArray(item, "recipeInstructions", "instructions"); ok {
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
}

func parseRecipeRating(item *microdata.Item, r *model.Recipe) {
	rating, ok := item.GetNestedItem("aggregateRating")
	if !ok {
		return
	}

	r.Rating = &model.AggregateRating{}
	if val, ok := getPropertyInt(rating, "ratingCount"); ok {
		r.Rating.RatingCount = val
	}
	if val, ok := getPropertyFloat(rating, "ratingValue"); ok {
		r.Rating.RatingValue = val
	}
	if val, ok := getPropertyInt(rating, "bestRating"); ok {
		r.Rating.BestRating = val
	}
	if val, ok := getPropertyInt(rating, "worstRating"); ok {
		r.Rating.WorstRating = val
	}
	if val, ok := getPropertyInt(rating, "reviewCount"); ok {
		r.Rating.ReviewCount = val
	}
}

func parseRecipeVideo(item *microdata.Item, r *model.Recipe, baseUrl *url.URL) {
	videoItem, ok := item.GetNestedItem("video")
	if !ok {
		return
	}

	video := &model.VideoObject{}
	if val, ok := getPropertyString(videoItem, "name"); ok {
		video.Name = utils.CleanupInline(val)
	}
	if val, ok := getPropertyString(videoItem, "description"); ok {
		video.Description = utils.CleanupInline(val)
	}
	if val, ok := getPropertyString(videoItem, "duration"); ok {
		video.Duration = utils.CleanupInline(val)
	}
	if val, ok := getPropertyString(videoItem, "embedUrl", "embedURL", "url"); ok {
		video.EmbedUrl = utils.ToAbsoluteUrl(baseUrl, val)
	}
	if val, ok := getPropertyString(videoItem, "contentURL", "contentUrl"); ok {
		video.ContentUrl = utils.ToAbsoluteUrl(baseUrl, val)
	}
	if val, ok := getPropertyString(videoItem, "thumbnailUrl", "image"); ok {
		video.ThumbnailUrl = utils.ToAbsoluteUrl(baseUrl, val)
	}
	if val, ok := getPropertyString(videoItem, "uploadDate", "datePublished"); ok {
		video.UploadDate, _ = utils.ParseRFC3339(val)
	}
	r.Video = video
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
				for _, child := range nested.Items {
					if child.IsOfSchemaType("Recipe") {
						recipe := &model.Recipe{Publisher: publisher}
						parseRecipe(child, recipe, baseUrl)
						feed.AddEntry(recipe)
					}
				}
			}
		}

		// Many recipes on the page itself
		if item.IsOfSchemaType("Recipe") {
			// A recipe that consists of another recipes, case of https://foodnetwork.co.uk/collections/air-fryer-recipes
			if nested, ok := item.GetNested("hasPart"); ok {
				for _, child := range nested.Items {
					if child.IsOfSchemaType("Recipe") {
						recipe := &model.Recipe{Publisher: publisher}
						parseRecipe(child, recipe, baseUrl)
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
