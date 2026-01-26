package custom_test

import (
	"testing"

	"github.com/borschtapp/krip"
	"github.com/stretchr/testify/assert"
)

func TestGoustoOnline(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping online test in short mode")
	}

	var website = "https://www.gousto.co.uk/cookbook/beef-recipes/simply-perfect-beef-spag-bol"
	recipe, err := krip.ScrapeUrl(website)
	assert.NoError(t, err)
	assert.True(t, recipe.IsValid())
	t.Log(recipe.String())
}
