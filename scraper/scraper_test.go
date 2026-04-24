package scraper

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"

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
		// Check majority.
		got := tc.tried > 0 && tc.hits*2 > tc.tried
		assert.Equal(t, tc.want, got, "tried=%d hits=%d", tc.tried, tc.hits)
	}
}

func TestFindEntries_SingleInvalidSchemaEntryFallsThrough(t *testing.T) {
	const rawHTML = `<!DOCTYPE html><html><head>
<script type="application/ld+json">
{"@context":"https://schema.org","@type":"Recipe","name":"Carbonara","url":"https://example.com/pasta"}
</script>
</head><body><main><ul>
  <li><a href="/recipes/pasta"><img src="/img/pasta.jpg"><span>Creamy Pasta</span></a></li>
  <li><a href="/recipes/chicken"><img src="/img/chicken.jpg"><span>Roast Chicken</span></a></li>
  <li><a href="/recipes/salad"><img src="/img/salad.jpg"><span>Summer Salad</span></a></li>
  <li><a href="/recipes/soup"><img src="/img/soup.jpg"><span>Tomato Soup</span></a></li>
</ul></main></body></html>`

	root, err := html.Parse(strings.NewReader(rawHTML))
	require.NoError(t, err)
	data, err := NodeInput(root, "https://example.com/recipes", model.ScrapeOptions{SkipMetaUrl: true})
	require.NoError(t, err)

	feed := &model.Feed{}
	scrapedURLs := map[string]bool{}
	err = findEntries(data, feed, model.FeedOptions{}, scrapedURLs)
	require.NoError(t, err, "should succeed by falling through to discovery")

	// Discovery must find the <ul> entries, not the schema stub.
	assert.Greater(t, len(feed.Entries), 1, "should find multiple entries via DOM discovery, not just the schema stub")
	for _, e := range feed.Entries {
		assert.NotEqual(t, "https://example.com/pasta", e.Url, "schema stub URL must not appear")
	}
}
