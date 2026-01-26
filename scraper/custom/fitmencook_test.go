package custom_test

import (
	"testing"

	"github.com/borschtapp/krip"
	"github.com/stretchr/testify/assert"
)

func TestGitMenCookOnline(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping online test in short mode")
	}

	var website = "https://fitmencook.com/recipes/healthy-chili-recipe/"
	recipe, err := krip.ScrapeUrl(website)
	assert.NoError(t, err)
	assert.True(t, recipe.IsValid())
	t.Log(recipe.String())
}
