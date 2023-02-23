package opengraph

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"

	"github.com/borschtapp/krip/model"
)

func TestOpenGraphParser(t *testing.T) {
	page := `
<html>
<head>
<meta property="og:locale" content="en_US" />
<meta property="og:title" content="Rapid Stir-Fried Chinese Beef" />
<meta property="og:url" content="https://www.hellofresh.com/recipes/uk-stir-fried-chinese-beef-5845b40b2e69d7259304d962" />
<meta property="og:image" content="https://img.hellofresh.com/f_auto,fl_lossy,h_640,q_auto,w_1200/hellofresh_s3/image/uk-stir-friend-chinese-beef-b5fd1d10.jpg" />
<meta property="og:description" content="In this recipe, we’re tossing the classic combo of beef and broccoli with bouncy noodles and dressing them in a savory soy and hoisin-based sauce." />
</head>
</html>
`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(page))
	assert.NoError(t, err)
	input := model.DataInput{Document: doc}

	recipe := &model.Recipe{}
	recipe.Author = &model.Person{}
	recipe.Publisher = &model.Organization{}
	assert.NoError(t, Scrape(&input, recipe))

	assert.Equal(t, "Rapid Stir-Fried Chinese Beef", recipe.Name)
	assert.Equal(t, "In this recipe, we’re tossing the classic combo of beef and broccoli with bouncy noodles and dressing them in a savory soy and hoisin-based sauce.", recipe.Description)
	assert.NotEmpty(t, recipe.Images)
	assert.Equal(t, "en-US", recipe.Language)
}
