package main

import (
	"flag"
	"fmt"
	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/scraper"
	"golang.org/x/net/html/charset"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/borschtapp/krip"
	"github.com/borschtapp/krip/testdata"
	"github.com/borschtapp/krip/utils"
)

func main() {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage of %s [options]:\n", os.Args[0])
		flag.PrintDefaults()
		_, _ = fmt.Fprint(os.Stderr, "\nUpdate all testdata's website HTML sources. Use to automate some routine.\n")
	}
	flag.Parse()

	_ = filepath.Walk(testdata.WebsitesDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(info.Name(), testdata.HtmlExt) {
			input, err := scraper.FileInput(path, model.InputOptions{SkipSchema: true})
			if err != nil {
				log.Fatal("Unable to read old testdata: " + err.Error())
			}

			fmt.Println("Saving " + input.Url)
			updateTestdata(input.Url)
		}
		return nil
	})
}

func updateTestdata(url string) {
	alias := utils.HostAlias(url)
	websiteFileName := testdata.WebsitesDir + alias + testdata.HtmlExt
	recipeFileName := testdata.RecipesDir + alias + testdata.JsonExt

	resp, err := http.Get(url)
	if err != nil {
		log.Println("Unable to fetch content: " + err.Error())
		return
	}
	if resp != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		log.Println("Bad response status: " + resp.Status)
		return
	}

	contentType := resp.Header.Get("Content-Type")
	reader, err := charset.NewReader(resp.Body, contentType)
	if err != nil {
		log.Println("Unable to read the page: " + err.Error())
		return
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		log.Println("Unable to read the content: " + err.Error())
		return
	}

	if err = os.WriteFile(websiteFileName, content, 0644); err != nil {
		log.Println("Unable to create file: " + err.Error())
		return
	}

	recipe, err := krip.ScrapeFile(websiteFileName)
	if err != nil {
		log.Println("Unable to scrape recipe: " + err.Error())
		return
	}

	if err = os.WriteFile(recipeFileName, []byte(recipe.String()), 0644); err != nil {
		log.Println("Unable to create recipe file: " + err.Error())
		return
	}
}
