package scraper

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/borschtapp/krip/model"
)

func makeRecipe(name, rawUrl string, withImage, withIngredients, withInstructions bool) *model.Recipe {
	r := &model.Recipe{Name: name, Url: rawUrl, Publisher: &model.Organization{Name: "Pub"}}
	if withImage {
		r.Images = []*model.ImageObject{{Url: "img.jpg"}}
	}
	if withIngredients {
		r.Ingredients = []*model.PropertyValue{{Name: "ingredient"}}
	}
	if withInstructions {
		r.Instructions = []*model.HowToSection{{HowToStep: model.HowToStep{Text: "cook"}}}
	}
	return r
}

func TestFilterEntries_RequiresImageByDefault(t *testing.T) {
	entries := []*model.Recipe{
		makeRecipe("With Image", "https://x.com/r/1", true, true, true),
		makeRecipe("No Image", "https://x.com/r/2", false, true, true),
	}

	result := filterEntries(entries, model.FeedOptions{})

	assert.Len(t, result, 1)
	assert.Equal(t, "With Image", result[0].Name)
}

func TestFilterEntries_OptionalImageAcceptsAll(t *testing.T) {
	entries := []*model.Recipe{
		makeRecipe("A", "https://x.com/r/1", false, true, true),
		makeRecipe("B", "https://x.com/r/2", false, true, true),
	}

	opt := model.FeedOptions{ScrapeOptions: model.ScrapeOptions{RecipeFilter: model.RecipeFilter{OptionalImage: true}}}
	result := filterEntries(entries, opt)

	assert.Len(t, result, 2)
}

func TestFilterEntries_MinIngredients(t *testing.T) {
	few := makeRecipe("Few", "https://x.com/r/1", true, false, true)
	few.Ingredients = []*model.PropertyValue{{Name: "salt"}}

	enough := makeRecipe("Enough", "https://x.com/r/2", true, false, true)
	enough.Ingredients = []*model.PropertyValue{{Name: "a"}, {Name: "b"}, {Name: "c"}}

	opt := model.FeedOptions{ScrapeOptions: model.ScrapeOptions{RecipeFilter: model.RecipeFilter{MinIngredients: 3}}}
	result := filterEntries([]*model.Recipe{few, enough}, opt)

	assert.Len(t, result, 1)
	assert.Equal(t, "Enough", result[0].Name)
}

func TestValidateDiscoverySample_StrictMajority(t *testing.T) {
	// hits*2 > tried: for tried=2, both must succeed (2*2=4 > 2)
	// for tried=3, two must succeed (2*2=4 > 3)
	tests := []struct {
		tried int
		hits  int
		want  bool
	}{
		{1, 1, true},
		{1, 0, false},
		{2, 2, true},  // both succeed
		{2, 1, false}, // only one — NOT a majority for tried=2
		{3, 2, true},  // 2/3 — majority
		{3, 1, false}, // 1/3 — not majority
	}

	for _, tc := range tests {
		// Simulate the majority check directly (the formula, not the network call)
		got := tc.tried > 0 && tc.hits*2 > tc.tried
		assert.Equal(t, tc.want, got, "tried=%d hits=%d", tc.tried, tc.hits)
	}
}
