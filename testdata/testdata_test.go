package testdata

import (
	"github.com/borschtapp/krip"
	"github.com/borschtapp/krip/utils"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/scraper"
)

func TestFilenames(t *testing.T) {
	WalkTestdataWebsites(func(name string, path string) {
		t.Run(name, func(t *testing.T) {
			input, err := scraper.FileInput(path, model.InputOptions{SkipText: true, SkipSchema: true})
			assert.NoError(t, err)
			assert.NotEmpty(t, input.Url)

			expected := utils.HostAlias(input.Url)
			assert.NotRegexp(t, regexp.MustCompile(`^file://.+`), input.Url)
			assert.Equal(t, expected+HtmlExt, name, "Incorrect filename for "+input.Url)
		})
	})
}

func TestWebsiteFiles(t *testing.T) {
	MockRequests(t)
	t.Parallel()

	WalkTestdataWebsites(func(name string, path string) {
		t.Run(name, func(t *testing.T) {
			recipe, err := krip.ScrapeFile(path)
			assert.NoError(t, err)

			AssertJson(t, recipe, RecipesDir+strings.TrimSuffix(name, HtmlExt))
		})
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
func TestWebsitesOnline(t *testing.T) {
	t.Skip("Skip online tests")
	t.Parallel()

	WalkTestdataWebsites(func(name string, path string) {
		t.Run(name, func(t *testing.T) {
			input, err := scraper.FileInput(path, model.InputOptions{SkipSchema: true})
			assert.NoError(t, err)
			assert.NotEmpty(t, input.Url)

			recipe, err := krip.ScrapeUrl(input.Url)
			assert.NoError(t, err)

			AssertJson(t, recipe, RecipesDir+strings.TrimSuffix(name, HtmlExt))
		})
	})
}

func TestWebsitesProperties(t *testing.T) {
	t.Parallel()
	t.Skip("TODO: doesn't work for all")

	var domains []string
	WalkTestdataRecipes(func(name string, recipe model.Recipe) {
		t.Run(name, func(t *testing.T) {
			assert.NotEmpty(t, recipe.Url)
			assert.NotEmpty(t, recipe.Name)
			assert.NotEmpty(t, recipe.Images)
			assert.NotEmpty(t, recipe.Ingredients)
			assert.NotEmpty(t, recipe.Instructions)
			assert.NotEmpty(t, recipe.Publisher)

			if !t.Failed() {
				domains = append(domains, strings.ReplaceAll(strings.ReplaceAll(recipe.Publisher.Url, "http://", "https://"), "//www.", "//"))
			}
		})
	})

	assert.NoError(t, updateReadme(domains))
}

func updateReadme(domains []string) error {
	readmeContent, err := os.ReadFile(PackageDir + "../README.md")
	if err != nil {
		return err
	}

	newReadme := ""
	for _, line := range strings.Split(string(readmeContent), "\n") {
		newReadme += line + "\n"

		if line == "[//]: # (This list is generated automatically, do not edit manually)" {
			break
		}
	}

	sort.Strings(domains)
	for _, domain := range domains {
		newReadme += "- " + domain + "\n"
	}

	if string(readmeContent) != newReadme {
		return os.WriteFile(PackageDir+"../README.md", []byte(newReadme), 0644)
	}
	return nil
}
