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

func TestSingleWebsite(t *testing.T) {
	t.Skip("Skip online tests")

	// var website = "https://www.thepioneerwoman.com/food-cooking/recipes/a11059/restaurant-style-salsa/"
	// recipe, err := ScrapeUrl(website)

	var website = "kitchenstories"
	recipe, err := krip.ScrapeFile(WebsitesDir + website + HtmlExt)
	assert.NoError(t, err)

	AssertJson(t, recipe, RecipesDir+website)
}

func TestTestdataWebsites(t *testing.T) {
	MockRequests(t)
	t.Parallel()

	_ = filepath.Walk(WebsitesDir, func(path string, info os.FileInfo, err error) error {
		assert.NoError(t, err)

		if !info.IsDir() && strings.HasSuffix(info.Name(), HtmlExt) {
			t.Run(info.Name(), func(t *testing.T) {
				recipe, err := krip.ScrapeFile(path)
				assert.NoError(t, err)

				AssertJson(t, recipe, RecipesDir+strings.TrimSuffix(info.Name(), HtmlExt))
			})
		}
		return nil
	})
}

func TestTestdataWebsitesOnline(t *testing.T) {
	t.Skip("Skip online tests")
	t.Parallel()

	_ = filepath.Walk(WebsitesDir, func(path string, info os.FileInfo, err error) error {
		assert.NoError(t, err)

		if !info.IsDir() && strings.HasSuffix(info.Name(), HtmlExt) {
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
