# Krip* 🇺🇦

<p align="center">
    <a href="https://pkg.go.dev/github.com/borschtapp/krip?tab=doc"><img src="https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white" alt="godoc" title="godoc"/></a>
    <a href="https://github.com/borschtapp/krip/tags"><img src="https://img.shields.io/github/v/tag/borschtapp/krip" alt="semver tag" title="semver tag"/></a>
    <a href="https://goreportcard.com/report/github.com/borschtapp/krip"><img src="https://goreportcard.com/badge/github.com/borschtapp/krip" alt="go report card" title="go report card"/></a>
    <a href="https://github.com/borschtapp/krip/blob/main/LICENSE"><img src="https://img.shields.io/github/license/borschtapp/krip" alt="license" title="license"/></a>
</p>

A Go library for fast, comprehensive and generalised scraping of culinary recipes from any website or HTML file.

* _Krip_ is a Ukrainian word for _dill_. The bud of the dill looks like web pages connected by a stem (a reference to the web and graphs).

---

I started this project as I wanted to build my own recipe keeper and found that there is only one
library that everyone uses for scraping recipes [recipe-scrapers](https://github.com/hhursev/recipe-scrapers/) written in Python.
The library is great, but contains lots of scrappers that seem redundant, however, it wasn't able to scrape the recipes I wanted.
I sent some PullRequests to it, but the language I chose for the project is Go, so I decided to rewrite it in Go.

This library contains completely rewritten parsers that are slightly inspired by the Python library.
I focused on speed and flexibility to cover most of the possible schemas and websites from the beginning and to retrieve a rich model.
Still, it supports per-domain customisation in case someone does not use a schema.

_Note:_ WIP, I am still learning how to use Go. Any hint or advice is welcome.

## Install
```
go get -u github.com/borschtapp/krip
```

## Features
- Parses microdata, OpenGraph and json-ld schemas
- Corrects erroneous JSON in the source code of websites (e.g. `jsonc` with comments or new lines)
- The resulting `Recipe` struct (object) is compatible with the [https://schema.org/Recipe](https://schema.org/Recipe) schema (see [comments](model/recipe.go))
- Includes custom parsers for specific websites (domains) that do not use any known recipe schema
- Removes empty, duplicate values and performs some normalization
- Command-line tool to scrape recipes from the internet
- Fast and efficient, thanks Go :)

## Contributing
Contributions are welcome! Please open an issue or a PR if you have any ideas or found a bug.\
Most common way to contribute is to add a custom parser for a specific website (if you have trouble, please open an issue and I will help you).

## Implementing custom scrapers
All you need is to implement a [`Scraper`](model/scraper.go) interface and register it via `krip.RegisterScraper()`.

Take a look at the already implemented custom scrapers:
- [Custom scraper for `https://fitmencook.com/`](scraper/website/fitmencook.go)
- [Custom scraper for `https://marleyspoon.com/`](scraper/website/marleyspoon.go)
- [Custom scraper for `https://kitchenstories.com/`](scraper/website/kitchenstories.go)


### To-Do List
- [ ] more custom parsers, implement all from the python library
- [ ] allergens support (missing in recipe schema)
- [ ] parsing of ingredients (missing in recipe schema)
- [ ] parsing of recipes from text
- [ ] validation and normalisation of units, constants, etc.

## Usage

### Command-line tool
```bash
go install github.com/borschtapp/krip/cmd/krip
krip --help
krip https://cooking.nytimes.com/recipes/3783-original-plum-torte
```

### Scrape recipe from web
```go
recipe, err := krip.ScrapeUrl("https://cooking.nytimes.com/recipes/3783-original-plum-torte")
if err != nil {
  // handle err
}

// Retrieve the recipe data
name := recipe.Name
ingredients := recipe.Ingredients
instructions := recipe.Instructions

// Print the recipe as JSON
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
- https://acouplecooks.com
- https://allrecipes.com
- https://alltommat.expressen.se
- https://altonbrown.com
- https://amazingribs.com
- https://ambitiouskitchen.com
- https://archanaskitchen.com
- https://arla.se
- https://aspicyperspective.com
- https://atelierdeschefs.fr
- https://averiecooks.com
- https://baking-sense.com
- https://bakingmischief.com
- https://bbc.co.uk/food
- https://bbcgoodfood.com
- https://bettybossi.ch
- https://bettycrocker.com
- https://biancazapatka.com/en
- https://bigoven.com
- https://blueapron.com
- https://bodybuilding.com
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
- https://cookeatshare.com
- https://cookieandkate.com
- https://cookincanuck.com
- https://cooking.nytimes.com
- https://cookinglight.com
- https://cookpad.com
- https://cookstr.com
- https://copykat.com
- https://countryliving.com
- https://creativecanning.com
- https://cucchiaio.it
- https://cuisineaz.com
- https://cybercook.com.br
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
- https://eattolerant.de
- https://eatwhattonight.com
- https://elanaspantry.com
- https://emmikochteinfach.de
- https://epicurious.com
- https://ethanchlebowski.com
- https://feastingathome.com
- https://fifteenspatulas.com
- https://food.com
- https://food52.com
- https://foodandwine.com
- https://foodinjars.com
- https://foodrepublic.com
- https://forksoverknives.com
- https://forktospoon.com
- https://framedcooks.com
- https://franzoesischkochen.de
- https://giallozafferano.it
- https://gimmedelicious.com
- https://gimmesomeoven.com
- https://gonnawantseconds.com
- https://gousto.co.uk
- https://greatbritishchefs.com
- https://halfbakedharvest.com
- https://handletheheat.com
- https://hassanchef.com
- https://headbangerskitchen.com
- https://healthy-delicious.com
- https://heb.com
- https://hellofresh.co.uk
- https://homechef.com
- https://homesicktexan.com
- https://hostthetoast.com
- https://ibreatheimhungry.com
- https://ica.se
- https://iheartrecipes.com
- https://im-worthy.com
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
- https://leanandgreenrecipes.net
- https://lecker.de
- https://lecremedelacrumb.com
- https://lekkerensimpel.com
- https://littlespicejar.com
- https://livelytable.com
- https://lovingitvegan.com
- https://lowcarbmaven.com
- https://maangchi.com
- https://madensverden.dk
- https://madewithlau.com
- https://marleyspoon.de
- https://marmiton.org
- https://marthastewart.com
- https://matprat.no
- https://melskitchencafe.com
- https://minimalistbaker.com
- https://misya.info/misya-srl-unipersonale
- https://mob.co.uk
- https://mobile.kptncook.com
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
- https://pressureluckcooking.com
- https://primaledgehealth.com
- https://przepisy.pl
- https://purelypope.com
- https://purplecarrot.com
- https://rachlmansfield.com
- https://rainbowplantlife.com
- https://realsimple.com
- https://receitas.globo.com
- https://recipes.timesofindia.com
- https://recipetineats.com
- https://redhousespice.com
- https://ruled.me
- https://rutgerbakt.nl
- https://sallysbakingaddiction.com
- https://saveur.com
- https://seasonsandsuppers.ca
- https://seriouseats.com
- https://simple-veganista.com
- https://simply-cookit.com
- https://simplyquinoa.com
- https://simplyrecipes.com
- https://simplywhisked.com
- https://skinnytaste.com
- https://smulweb.nl
- https://southerncastiron.com
- https://southernliving.com
- https://spendwithpennies.com
- https://springlane.de
- https://steamykitchen.com
- https://sunbasket.com
- https://sundpaabudget.dk
- https://sunset.com
- https://sweetcsdesigns.com
- https://sweetpeasandsaffron.com
- https://tasteaholics.gladness.in
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
- https://thevintagemixer.com
- https://thewoksoflife.com
- https://tine.no
- https://tudogostoso.com.br
- https://twopeasandtheirpod.com
- https://valdemarsro.dk
- https://vanillaandbean.com
- https://vegolosi.it
- https://vegrecipesofindia.com
- https://watchwhatueat.com
- https://weightwatchers.com
- https://whatsgabycooking.com
- https://wholefully.com
- https://yemek.com
- https://yummly.com
- https://zenbelly.com
