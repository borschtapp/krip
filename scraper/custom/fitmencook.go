package custom

import (
	"github.com/PuerkitoBio/goquery"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/utils"
)

func ScrapeFitMenCook(data *model.DataInput, r *model.Recipe) error {
	if data.Document == nil {
		return nil
	}

	data.Document.Find(".fmc_ingredients ul li").Each(func(_ int, s *goquery.Selection) {
		if s.Has("strong").Length() > 0 {
			return // skip section headers
		}
		if text := utils.CleanupInline(s.Text()); text != "" {
			r.Ingredients = append(r.Ingredients, &model.PropertyValue{Name: text})
		}
	})

	data.Document.Find(".fmc_recipe_steps .fmc_step_content").Each(func(_ int, s *goquery.Selection) {
		if text := utils.CleanupInline(s.Text()); text != "" {
			r.Instructions = append(r.Instructions, &model.HowToSection{HowToStep: model.HowToStep{Text: text}})
		}
	})

	return nil
}
