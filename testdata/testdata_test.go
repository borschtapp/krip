package testdata

import (
	"github.com/borschtapp/krip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/scraper"
)

func TestTestdataWebsites(t *testing.T) {
	MockRequests(t)
	t.Parallel()

	_ = filepath.Walk(WebsitesDir, func(path string, info os.FileInfo, err error) error {
		assert.NoError(t, err)

		if strings.HasSuffix(info.Name(), HtmlExt) {
			t.Run(info.Name(), func(t *testing.T) {
				recipe, err := krip.ScrapeFile(path)
				assert.NoError(t, err)

				AssertJson(t, recipe, RecipesDir+strings.TrimSuffix(info.Name(), HtmlExt))
			})
		}
		return nil
	})
}

/*
The below list of hosts had `403 Forbidden` status last time due to Cloudflare or so
https://www.blueapron.com/recipes/bbq-chickpeas-farro-with-corn-cucumbers-hard-boiled-eggs-3
http://www.bunkycooks.com/2011/12/the-best-three-cheese-lasagna-recipe/
https://dinnerthendessert.com/indian-chicken-korma/
https://downshiftology.com/recipes/greek-chicken-kabobs/
https://www.homechef.com/meals/prosciutto-and-mushroom-carbonara-standard
https://www.latelierderoxane.com/blog/recette-cake-marbre/
https://www.marmiton.org/recettes/recette_ratatouille_23223.aspx
https://sundpaabudget.dk/one-pot-pasta-med-kyllingekebab/
https://www.thekitchn.com/manicotti-22949270
https://www.tudogostoso.com.br/receita/128825-caipirinha-original.html
https://www.heb.com/recipe/recipe-item/truffled-spaghetti-squash/1398755977632 (denied by browser visit too, down?)
https://healthyeating.nhlbi.nih.gov/recipedetail.aspx?cId=3&rId=188&AspxAutoDetectCookieSupport=1 (>10 redirects)
*/
func TestTestdataWebsitesOnline(t *testing.T) {
	t.Skip("Skip online tests")
	t.Parallel()

	_ = filepath.Walk(WebsitesDir, func(path string, info os.FileInfo, err error) error {
		assert.NoError(t, err)

		if strings.HasSuffix(info.Name(), HtmlExt) {
			t.Run(info.Name(), func(t *testing.T) {
				input, err := scraper.FileInput(path, model.InputOptions{SkipSchema: true})
				assert.NoError(t, err)
				assert.NotEmpty(t, input.Url)

				recipe, err := krip.ScrapeUrl(input.Url)
				assert.NoError(t, err)

				AssertJson(t, recipe, RecipesDir+strings.TrimSuffix(info.Name(), HtmlExt))
			})
		}
		return nil
	})
}
