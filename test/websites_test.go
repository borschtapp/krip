package test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/borschtapp/krip"
	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/scraper"
	"github.com/borschtapp/krip/utils"
)

func TestGetNoInstructions(t *testing.T) {
	MockRequests(t)
	t.Parallel()
	t.Skip("TODO: fix")

	var domains []string
	_ = filepath.Walk(WebsitesDir, func(path string, info os.FileInfo, err error) error {
		assert.NoError(t, err)

		if !info.IsDir() && strings.HasSuffix(info.Name(), HtmlExt) {
			t.Run(info.Name(), func(t *testing.T) {
				recipe, err := krip.ScrapeFile(path)
				assert.NoError(t, err)
				assert.NotEmpty(t, recipe.Url)
				assert.NotEmpty(t, recipe.Name)
				assert.NotEmpty(t, recipe.Description)
				assert.NotEmpty(t, recipe.Language)
				assert.True(t, len(recipe.ThumbnailUrl) > 0 || len(recipe.Images) > 0)

				//assert.NotEmpty(t, recipe.Yield)
				//assert.NotEmpty(t, recipe.Ingredients)
				//assert.NotEmpty(t, recipe.Instructions)
				//assert.NotEmpty(t, recipe.Publisher)

				if !t.Failed() {
					domains = append(domains, utils.Hostname(recipe.Url))
				}
			})
		}
		return nil
	})

	assert.NoError(t, updateReadme(domains))
}

func TestTestdataFilenames(t *testing.T) {
	_ = filepath.Walk(WebsitesDir, func(path string, info os.FileInfo, err error) error {
		assert.NoError(t, err)

		if strings.HasSuffix(info.Name(), HtmlExt) {
			t.Run(info.Name(), func(t *testing.T) {
				input, err := scraper.FileInput(path, model.InputOptions{SkipText: true, SkipSchema: true})
				assert.NoError(t, err)
				assert.NotEmpty(t, input.Url)

				expected := utils.HostAlias(input.Url)
				assert.NotRegexp(t, regexp.MustCompile(`^file://.+`), input.Url)
				assert.Equal(t, expected+HtmlExt, info.Name(), "Incorrect filename for "+input.Url)
			})
		}
		return nil
	})
}

func updateReadme(domains []string) error {
	readmeContent, err := os.ReadFile(TestdataDir + "../README.md")
	if err != nil {
		return err
	}

	newReadme := ""
	for _, line := range strings.Split(string(readmeContent), "\n") {
		newReadme += line + "\n"

		if line == "## Verified sources" {
			break
		}
	}

	sort.Strings(domains)
	for _, domain := range domains {
		newReadme += "- https://" + domain + "\n"
	}

	if string(readmeContent) != newReadme {
		return os.WriteFile(TestdataDir+"../README.md", []byte(newReadme), 0644)
	}
	return nil
}
