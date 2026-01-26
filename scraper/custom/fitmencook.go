package custom

import (
	"github.com/PuerkitoBio/goquery"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/utils"
)

func ScrapeFitMenCook(data *model.DataInput, r *model.Recipe) error {
	if data.Document != nil {
		if s := data.Document.Find(".fmc_ingredients ul li"); len(s.Nodes) != 0 {
			s.Each(func(i int, s *goquery.Selection) {
				if s1 := s.Has("strong"); len(s1.Nodes) != 0 {
					return
				}

				text := utils.CleanupInline(s.Text())
				if text != "" {
					r.Ingredients = append(r.Ingredients, text)
				}
			})
		}

		if s := data.Document.Find(".fmc_recipe_steps .fmc_step_content"); len(s.Nodes) != 0 {
			s.Each(func(i int, s *goquery.Selection) {
				text := utils.CleanupInline(s.Text())
				if text != "" {
					r.Instructions = append(r.Instructions, &model.HowToSection{HowToStep: model.HowToStep{Text: text}})
				}
			})
		}
	}

	return nil
}
