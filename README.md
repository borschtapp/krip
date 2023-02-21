# Krip - Quick, comprehensive and generalized recipes scraping

A Go library for scraping culinary recipes from any website or HTML file.

---

I found it illogical for Go users to use Python for scraping recipes.
The Python library [recipe-scrapers](https://github.com/hhursev/recipe-scrapers/) is great, but slow and very limited.

This library contains completely rewritten parsers that are slightly inspired by the Python library.
I focused on speed and flexibility to cover most of the possible schemas and websites out of the box and retrives extensive model.
Still, it supports per-domain customisation in case someone does not use any schema.

_Note:_ WIP, I'm still learning how to use Go. Found it fun, but a difficult to switch after OOP.

## Install
```
go get -u github.com/borschtapp/krip
```

## Features
- The resulting `Recipe` struct (object) is compatible with [Recipe schema](https://schema.org/Recipe) (see [comments](model/recipe.go))
- Scrapes any website that uses the recipe schema, even if it is broken
- Includes parsers for custom domains (sources) that don't use the schema
- Removes empty, duplicate values and performs some normalization on the fly
- Fast and efficient, thanks Go :)

### To-Do List
- [ ] more custom domain parsers, implement all from the python library
- [ ] allergens support (missing in recipe schema)
- [ ] parsing of ingredients (missing in recipe schema)
- [ ] parsing of recipes from text
- [ ] validation and normalisation of units, constants, etc.

## Usage

### Scrape recipe from web
```go
recipe, err := krip.ScrapeUrl("https://cooking.nytimes.com/recipes/3783-original-plum-torte")
fmt.Println(recipe)
```
```json
{
  "@id": "https://cooking.nytimes.com/recipes/3783-original-plum-torte",
  "name": "Original Plum Torte",
  "thumbnailUrl": "https://static01.nyt.com/images/2019/09/07/dining/plumtorte/plumtorte-articleLarge-v4.jpg",
  "author": {
    "name": "Marian Burros"
  },
  "publisher": {
    "name": "NYT Cooking",
    "url": "https://cooking.nytimes.com"
  },
  "inLanguage": "en-US",
  "description": "The Times published Marian Burros’s recipe for Plum Torte every September from 1983 until 1989, when the editors determined that enough was enough. The recipe was to be printed for the last time that year. “To counter anticipated protests,” Ms. Burros wrote a few years later, “the recipe was printed in larger type than usual with a broken-line border around it to encourage clipping.” It didn’t help. The paper was flooded with angry letters. “The appearance of the recipe, like the torte itself, is bittersweet,” wrote a reader in Tarrytown, N.Y. “Summer is leaving, fall is coming. That's what your annual recipe is all about. Don't be grumpy about it.” We are not! And we pledge that every year, as summer gives way to fall, we will make sure that the recipe is easily available to one and all. The original 1983 recipe called for 1 cup sugar; the 1989 version reduced that to 3/4 cup. We give both options below. Here are \u003ca href=\" http://www.nytimes.com/interactive/2016/09/14/dining/marian-burros-plum-torte-recipe-variations.html\"\u003efive ways to adapt the torte\u003c/a\u003e.",
  "totalTime": 75,
  "recipeCategory": [
    "breakfast",
    "brunch",
    "easy",
    "weekday",
    "times classics",
    "dessert"
  ],
  "keywords": [
    "flour",
    "plum",
    "unsalted butter",
    "nut-free",
    "vegetarian"
  ],
  "recipeYield": 8,
  "recipeIngredient": [
    "3/4 to 1 cup sugar",
    "1/2 cup unsalted butter, softened",
    "1 cup unbleached flour, sifted",
    "1 teaspoon baking powder",
    "Pinch of salt (optional)",
    "2 eggs",
    "24 halves pitted purple plums",
    "Sugar, lemon juice and cinnamon, for topping"
  ],
  "recipeInstructions": [
    {
      "text": "Heat oven to 350 degrees."
    },
    {
      "text": "Cream the sugar and butter in a bowl. Add the flour, baking powder, salt and eggs and beat well."
    },
    {
      "text": "Spoon the batter into a springform pan of 8, 9 or 10 inches. Place the plum halves skin side up on top of the batter. Sprinkle lightly with sugar and lemon juice, depending on the sweetness of the fruit. Sprinkle with about 1 teaspoon of cinnamon, depending on how much you like cinnamon."
    },
    {
      "text": "Bake 1 hour, approximately. Remove and cool; refrigerate or freeze if desired. Or cool to lukewarm and serve plain or with whipped cream. (To serve a torte that was frozen, defrost and reheat it briefly at 300 degrees.)"
    }
  ],
  "nutrition": {
    "calories": "350",
    "carbohydrateContent": "57 grams",
    "fatContent": "13 grams",
    "fiberContent": "3 grams",
    "proteinContent": "4 grams",
    "saturatedFatContent": "8 grams",
    "sodiumContent": "63 milligrams",
    "sugarContent": "42 grams",
    "transFatContent": "0 grams",
    "unsaturatedFatContent": "4 grams"
  },
  "aggregateRating": {
    "ratingCount": 8717,
    "ratingValue": 5
  }
}
```

## Tested on

The scraper contains a test for the source and was able to extract all the important fields, including but not limited to:
- `url`
- `name`
- `inLanguage`
- `thumbnailUrl`
- `recipeIngredient`
- `recipeInstructions`
- `publisher` (including `name` and `url`)

### For the following websites
[//]: # (This list is generated automatically, do not edit manually)
- https://101cookbooks.com
- https://750g.com
- https://Cookstr.com
- https://acouplecooks.com
- https://allrecipes.com
- https://alltommat.expressen.se
- https://amazingribs.com
- https://ambitiouskitchen.com
- https://archanaskitchen.com
- https://aspicyperspective.com
- https://atelierdeschefs.fr
- https://averiecooks.com
- https://baking-sense.com
- https://bakingmischief.com
- https://bbc.co.uk
- https://bettycrocker.com
- https://bigoven.com
- https://blueapron.com
- https://bonappetit.com
- https://bowlofdelicious.com
- https://budgetbytes.com
- https://castironketo.net
- https://cavemanketo.com
- https://cdkitchen.com
- https://chefkoch.de
- https://claudia.abril.com.br
- https://closetcooking.com
- https://comidinhasdochef.com
- https://cookieandkate.com
- https://cookincanuck.com
- https://cooking.nytimes.com
- https://cookinglight.com
- https://cookpad.com
- https://copykat.com
- https://countryliving.com
- https://creativecanning.com
- https://cucchiaio.it
- https://cuisineaz.com
- https://damndelicious.net
- https://davidlebovitz.com
- https://delish.com
- https://dinnerly.de
- https://dinnerthendessert.com
- https://ditchthecarbs.com
- https://domesticate-me.com
- https://downshiftology.com
- https://dr.dk
- https://eatingbirdfood.com
- https://eatingwell.com
- https://eatsmarter.com
- https://eatwhattonight.com
- https://elanaspantry.com
- https://epicurious.com
- https://fifteenspatulas.com
- https://food52.com
- https://foodandwine.com
- https://foodinjars.com
- https://foodrepublic.com
- https://forksoverknives.com
- https://framedcooks.com
- https://franzoesischkochen.de
- https://gimmedelicious.com
- https://gimmesomeoven.com
- https://gonnawantseconds.com
- https://greatbritishchefs.com
- https://halfbakedharvest.com
- https://hassanchef.com
- https://headbangerskitchen.com
- https://healthy-delicious.com
- https://heb.com
- https://hellofresh.co.uk
- https://homechef.com
- https://homesicktexan.com
- https://hostthetoast.com
- https://ibreatheimhungry.com
- https://iheartrecipes.com
- https://indianhealthyrecipes.com
- https://innit.com
- https://inspiralized.com
- https://jamieoliver.com
- https://jimcooksfoodgood.com
- https://joyfoodsunshine.com
- https://julieblanner.com
- https://justataste.com
- https://justonecookbook.com
- https://kennymcgovern.com
- https://ketoconnect.net
- https://ketogasm.com
- https://kingarthurbaking.com
- https://kitchenstories.com
- https://klopotenko.com
- https://knorr.com
- https://kochbar.de
- https://koket.se
- https://kuchnia-domowa.pl
- https://lecker.de
- https://lecremedelacrumb.com
- https://lekkerensimpel.com
- https://littlespicejar.com
- https://livelytable.com
- https://lovingitvegan.com
- https://lowcarbmaven.com
- https://madensverden.dk
- https://marleyspoon.de
- https://marmiton.org
- https://marthastewart.com
- https://matprat.no
- https://melskitchencafe.com
- https://minimalistbaker.com
- https://misya.info/misya-srl-unipersonale
- https://mob.co.uk
- https://momswithcrockpots.com
- https://motherthyme.com
- https://mybakingaddiction.com
- https://mykitchen101.com
- https://mykitchen101en.com
- https://myrecipes.com
- https://naturallyella.com
- https://nhs.uk
- https://ninjatestkitchen.eu
- https://nomnompaleo.com
- https://nourishedbynutrition.com
- https://ohsheglows.com
- https://ohsweetbasil.com
- https://omnivorescookbook.com
- https://paleorunningmomma.com
- https://picky-palate.com
- https://pinchofyum.com
- https://primaledgehealth.com
- https://przepisy.pl
- https://purelypope.com
- https://rachlmansfield.com
- https://rainbowplantlife.com
- https://realsimple.com
- https://receitas.globo.com
- https://recipes.timesofindia.com
- https://recipetineats.com
- https://redhousespice.com
- https://ricette.giallozafferano.it
- https://ruled.me
- https://sallysbakingaddiction.com
- https://saveur.com
- https://seasonsandsuppers.ca
- https://seriouseats.com
- https://simplyquinoa.com
- https://simplyrecipes.com
- https://skinnytaste.com
- https://southernliving.com
- https://spendwithpennies.com
- https://steamykitchen.com
- https://sunbasket.com
- https://sundpaabudget.dk
- https://sweetcsdesigns.com
- https://sweetpeasandsaffron.com
- https://tasteaholics.com
- https://tasteofhome.com
- https://tastesbetterfromscratch.com
- https://tastesoflizzyt.com
- https://tasty.co
- https://thatlowcarblife.com
- https://theblackpeppercorn.com
- https://thechunkychef.com
- https://theclevercarrot.com
- https://thehappyfoodie.co.uk
- https://thekitchenmagpie.com
- https://thekitchn.com
- https://thenutritiouskitchen.co
- https://thepioneerwoman.com
- https://therealfooddietitians.com
- https://thespruceeats.com
- https://thestayathomechef.com
- https://thewoksoflife.com
- https://tine.no
- https://tudogostoso.com.br
- https://twopeasandtheirpod.com
- https://valdemarsro.dk
- https://vanillaandbean.com
- https://vegolosi.it
- https://vegrecipesofindia.com
- https://watchwhatueat.com
- https://whatsgabycooking.com
- https://wholefully.com
- https://yemek.com
- https://yummly.com
- https://zenbelly.com
