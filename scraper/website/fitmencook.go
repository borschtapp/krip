package website

import (
	"github.com/PuerkitoBio/goquery"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/utils"
)

func ScrapeFitMenCook(data *model.DataInput, r *model.Recipe) error {
	if data.Document != nil {
		if s := data.Document.Find(".recipe-ingredients h4 strong").First(); len(s.Nodes) != 0 {
			if val := utils.FindNumber(s.Text()); val > 0 {
				r.Yield = int(val)
			}
		}

		if s := data.Document.Find("div.recipe-ingredients li"); len(s.Nodes) != 0 {
			s.Each(func(i int, s *goquery.Selection) {
				text := utils.CleanupInline(s.Text())
				if text != "" {
					r.Ingredients = append(r.Ingredients, text)
				}
			})
		}

		if s := data.Document.Find("div.recipe-steps > ol:first-of-type li"); len(s.Nodes) != 0 {
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
