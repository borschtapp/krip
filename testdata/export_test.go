package testdata

import (
	"github.com/borschtapp/krip/model"
	"os"
	"strings"
	"testing"
)

func TestCollectIngredients(t *testing.T) {
	t.Skip("irrelevant test, use by request only")
	t.Parallel()

	var ingredients []string
	WalkTestdataRecipes(func(name string, recipe model.Recipe) {
		t.Run(name, func(t *testing.T) {
			if recipe.Ingredients != nil {
				ingredients = append(ingredients, recipe.Ingredients...)
			}
		})
	})

	os.WriteFile(currentPath()+"/ingredients.txt", []byte(strings.Join(ingredients, "\n")), 0644)
}
