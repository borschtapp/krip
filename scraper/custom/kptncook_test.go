package custom_test

import (
	"testing"

	"github.com/borschtapp/krip"
	"github.com/stretchr/testify/assert"
)

func TestKptnCookOnline(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping online test in short mode")
	}

	var website = "https://mobile.kptncook.com/recipe/pinterest/Low-Carb-Flammkuchen%20mit%20Serranoschinken%20&amp;%20Frischk%C3%A4se/315c3c32"
	recipe, err := krip.ScrapeUrl(website)
	assert.NoError(t, err)
	assert.True(t, recipe.IsValid())
	t.Log(recipe.String())
}
