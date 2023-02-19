package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/html/charset"

	"github.com/borschtapp/krip"
	"github.com/borschtapp/krip/test"
	"github.com/borschtapp/krip/utils"
)

func main() {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage of %s [options] [url]:\n", os.Args[0])
		flag.PrintDefaults()
		_, _ = fmt.Fprint(os.Stderr, "\nAdds new recipe webpage to test catalog. Use to automate some routine.\n")
	}
	flag.Parse()

	switch len(flag.Args()) {
	case 1:
		recipeUrl := flag.Args()[0]
		createWebsiteTestdata(recipeUrl)
		fmt.Println("Done, test page added!")
	default:
		flag.Usage()
		os.Exit(1)
	}
}

func createWebsiteTestdata(recipeUrl string) {
	alias := utils.HostAlias(recipeUrl)
	websiteFileName := test.WebsitesDir + alias + test.HtmlExt
	recipeFileName := test.RecipesDir + alias + test.JsonExt

	if _, err := os.Stat(websiteFileName); err == nil {
		log.Fatal("Testdata already exists for the alias: " + alias)
	}

	resp, err := http.Get(recipeUrl)
	if err != nil {
		log.Fatal("Unable to fetch content: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatal("Bad response status: " + resp.Status)
	}

	contentType := resp.Header.Get("Content-Type")
	reader, err := charset.NewReader(resp.Body, contentType)
	if err != nil {
		log.Fatal("Unable to read the page: " + err.Error())
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		log.Fatal("Unable to read the content: " + err.Error())
	}

	if err = os.WriteFile(websiteFileName, content, 0644); err != nil {
		log.Fatal("Unable to create file: " + err.Error())
	}

	recipe, err := krip.ScrapeFile(websiteFileName)
	if err != nil {
		log.Fatal("Unable to scrape recipe: " + err.Error())
	}

	if err = os.WriteFile(recipeFileName, []byte(recipe.String()), 0644); err != nil {
		log.Fatal("Unable to create recipe file: " + err.Error())
	}
}
