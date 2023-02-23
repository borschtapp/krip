package website

import (
	"github.com/sosodev/duration"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/utils"
)

func ScrapeCookstr(data *model.DataInput, r *model.Recipe) error {
	if data.Document != nil {
		if text := utils.CleanupInline(data.Document.Find("h1.articleHeadline").First().Text()); text != "" {
			r.Name = text
		}

		if text := utils.CleanupInline(data.Document.Find(".articleDiv div.articleAttrSection:nth-child(6)").First().Text()); len(text) > 100 {
			r.Description = text
		}

		if text := utils.CleanupInline(data.Document.Find(".articleDiv div.articleAttrSection:nth-child(7)").First().Text()); len(text) > 100 {
			r.Text = text
		}

		if text := utils.CleanupInline(data.Document.Find("h4#commentCount").First().Text()); len(text) > 0 {
			r.CommentCount = int(utils.FindNumber(text))
		}

		data.Document.Find("span.attrLabel").Each(func(i int, s *goquery.Selection) {
			label := utils.CleanupInline(s.Text())
			value := utils.CleanupInline(strings.Replace(s.Parent().Text(), label, "", 1))

			if label == "Dietary Consideration" && value != "" {
				for _, diet := range strings.Split(value, ",") {
					r.Diets = utils.AppendUnique(r.Diets, utils.CleanupInline(diet))
				}
			} else if (label == "Makes" || label == "Serves") && value != "" {
				r.Yield = int(utils.FindNumber(value))
			} else if label == "Cooking Method" && value != "" {
				r.CookingMethod = value
			} else if label == "Type of Dish" && value != "" {
				r.Categories = strings.Split(value, ", ")
			} else if label == "Equipment" && value != "" {
				r.Equipment = strings.Split(value, ", ")
			} else if label == "Total Time" && value != "" {
				if val, ok := utils.ParseDuration(value); ok {
					r.TotalTime = duration.Format(val)
				}
			}
		})

		if len(r.Ingredients) == 0 {
			data.Document.Find("div.recipeIngredients li").Each(func(i int, s *goquery.Selection) {
				if text := utils.CleanupInline(s.Text()); text != "" {
					r.Ingredients = append(r.Ingredients, text)
				}
			})
		}

		if len(r.Instructions) == 0 {
			if s := data.Document.Find("div.stepByStepInstructionsDiv li"); len(s.Nodes) != 0 {
				s.Each(func(i int, s *goquery.Selection) {
					if text := utils.CleanupInline(s.Text()); text != "" {
						r.Instructions = append(r.Instructions, &model.HowToSection{HowToStep: model.HowToStep{Text: text}})
					}
				})
			}
		}
	}

	r.Publisher = &model.Organization{}
	r.Publisher.Name = "Cookstr.com"
	r.Publisher.Url = "https://www.cookstr.com"
	r.Publisher.Logo = "https://static.primecp.com/site_templates/3001/images/site_logo_sID66.png?v=22222"

	return nil
}
