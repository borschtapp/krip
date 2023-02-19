package krip

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/scraper"
	"github.com/borschtapp/krip/test"
)

func TestSingleWebsite(t *testing.T) {
	test.MockRequests(t)

	// var website = "https://www.thepioneerwoman.com/food-cooking/recipes/a11059/restaurant-style-salsa/"
	// recipe, err := ScrapeUrl(website)

	var website = "kitchenstories"
	recipe, err := ScrapeFile(test.WebsitesDir + website + test.HtmlExt)
	assert.NoError(t, err)

	test.AssertJson(t, recipe, test.RecipesDir+website)
}

func TestTestdataWebsites(t *testing.T) {
	test.MockRequests(t)
	t.Parallel()

	_ = filepath.Walk(test.WebsitesDir, func(path string, info os.FileInfo, err error) error {
		assert.NoError(t, err)

		if !info.IsDir() && strings.HasSuffix(info.Name(), test.HtmlExt) {
			t.Run(info.Name(), func(t *testing.T) {
				recipe, err := ScrapeFile(path)
				assert.NoError(t, err)

				test.AssertJson(t, recipe, test.RecipesDir+strings.TrimSuffix(info.Name(), test.HtmlExt))
			})
		}
		return nil
	})
}

func TestTestdataWebsitesOnline(t *testing.T) {
	t.Skip("Skip online tests")
	t.Parallel()

	_ = filepath.Walk(test.WebsitesDir, func(path string, info os.FileInfo, err error) error {
		assert.NoError(t, err)

		if !info.IsDir() && strings.HasSuffix(info.Name(), test.HtmlExt) {
			t.Run(info.Name(), func(t *testing.T) {
				input, err := scraper.FileInput(path, model.InputOptions{SkipSchema: true})
				assert.NoError(t, err)
				assert.NotEmpty(t, input.Url)

				recipe, err := ScrapeUrl(input.Url)
				assert.NoError(t, err)

				test.AssertJson(t, recipe, test.RecipesDir+strings.TrimSuffix(info.Name(), test.HtmlExt))
			})
		}
		return nil
	})
}
