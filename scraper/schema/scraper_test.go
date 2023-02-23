package schema

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"

	"github.com/astappiev/microdata"
	"github.com/borschtapp/krip/model"
)

func TestSchemaParser(t *testing.T) {
	page := `<script type="application/ld+json" id="schema-org">{"@context":"http://schema.org/","@type":"Recipe","name":"Rapid Stir-Fried Beef and Broccoli","author":"HelloFresh","image":"https://img.hellofresh.com/f_auto,fl_lossy,h_640,q_auto,w_1200/hellofresh_s3/image/uk-stir-friend-chinese-beef-b5fd1d10.jpg","thumbnailUrl":"https://img.hellofresh.com/f_auto,fl_lossy,h_300,q_auto,w_450/hellofresh_s3/image/uk-stir-friend-chinese-beef-b5fd1d10.jpg","description":"One of the great things about Asian-style stir-fries is that they deliver maximum flavor in a snap. In this recipe, we’re tossing the classic combo of beef and broccoli with bouncy noodles and dressing them in a savory soy and hoisin-based sauce. It comes together so swiftly in the pan, it might even be easier than picking up the phone for takeout!","datePublished":"2016-12-05T18:38:03+00:00","totalTime":"PT20M","nutrition":{"@type":"NutritionInformation","calories":"754 kcal","fatContent":"29 g","saturatedFatContent":"6 g","carbohydrateContent":"77 g","sugarContent":"9 g","proteinContent":"54 g","fiberContent":"6 g","cholesterolContent":"102 mg","sodiumContent":"1620 mg"},"recipeInstructions":[{"@type":"HowToStep","text":"Wash and dry all produce. Bring a large pot of salted water to a boil. Trim and thinly slice scallions. Mince or grate garlic. Peel and mince ginger. Whisk together sesame oil, 1 TBSP ketchup, soy sauce, 1½ TBSP hoisin sauce, and 1 TBSP water in a small bowl."},{"@type":"HowToStep","text":"Add broccoli to boiling water and cook until tender but still crisp, 3-4 minutes. Drain and rinse under cold water. Set aside."},{"@type":"HowToStep","text":"Toss steak tips with cornstarch in a large bowl. Season generously with salt and pepper. Heat a large drizzle of oil in a large pan over high heat. (TIP: If you have a nonstick pan, break it out.) Toss in steak tips and cook to desired doneness, 3-4 minutes. Remove and set aside."},{"@type":"HowToStep","text":"Heat a drizzle of oil in same pan over medium heat. Add garlic, ginger, and scallions and cook until fragrant, 1 minute, tossing. Toss in half the noodles from the package (we sent more) and a drizzle of oil. Break up noodles until they no longer stick together, using tongs or two wooden spoons."},{"@type":"HowToStep","text":"Pour in 1 cup water, cover, and steam until noodles are tender, 3 minutes. (TIP: If your pan doesn’t have a lid, carefully cover it with aluminum foil.) Uncover, increase heat to medium-high, and toss until noodles are tender, 3-4 minutes. Add sauce and toss to coat. Cook until sauce is thickened, 1 minute."},{"@type":"HowToStep","text":"Toss broccoli and steak into noodles to warm through. Season with as much sriracha as you like (careful, it’s spicy). Season with salt and pepper. Divide between plates and serve."}],"recipeIngredient":["12 ounce Beef Sirloin Tips","2 unit Scallions","2 clove Garlic","1 tablespoon Cornstarch","1 thumb Ginger","16 ounce Yakisoba Noodles","1 unit Ketchup","4 unit Soy Sauce","1 jar Hoisin Sauce Jar","8 ounce Broccoli Florets","1 tablespoon Sesame Oil","1 teaspoon Sriracha","4 teaspoon Vegetable Oil","unit Salt","unit Pepper"],"recipeYield":2,"keywords":["Rapid","Spicy"],"recipeCategory":"main course","recipeCuisine":"Asian"}</script>`

	data, err := microdata.ParseHTML(strings.NewReader(page), "", "https://www.hellofresh.com/recipes/uk-stir-fried-chinese-beef-5845b40b2e69d7259304d962")
	assert.NoError(t, err)

	input := model.DataInput{Schemas: data}
	assert.NotEmpty(t, input.Schemas)

	recipe := &model.Recipe{}
	recipe.Author = &model.Person{}
	recipe.Publisher = &model.Organization{}
	assert.NoError(t, Scrape(&input, recipe))

	assert.Equal(t, "Rapid Stir-Fried Beef and Broccoli", recipe.Name)
	assert.NotEmpty(t, recipe.Images)
	assert.Equal(t, "PT20M", recipe.TotalTime)
	assert.Equal(t, []string{"Rapid", "Spicy"}, recipe.Keywords)
	assert.Equal(t, []string{"main course"}, recipe.Categories)
	assert.Equal(t, []string{"Asian"}, recipe.Cuisines)
	assert.Equal(t, 2, recipe.Yield)

	assert.Len(t, recipe.Instructions, 6)
	assert.Len(t, recipe.Ingredients, 15)
	assert.Equal(t, []string{"12 ounce Beef Sirloin Tips",
		"2 unit Scallions",
		"2 clove Garlic",
		"1 tablespoon Cornstarch",
		"1 thumb Ginger",
		"16 ounce Yakisoba Noodles",
		"1 unit Ketchup",
		"4 unit Soy Sauce",
		"1 jar Hoisin Sauce Jar",
		"8 ounce Broccoli Florets",
		"1 tablespoon Sesame Oil",
		"1 teaspoon Sriracha",
		"4 teaspoon Vegetable Oil",
		"unit Salt",
		"unit Pepper"}, recipe.Ingredients)
}
