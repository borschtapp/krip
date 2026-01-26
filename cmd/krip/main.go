package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/borschtapp/krip"
)

func main() {
	feedFlag := flag.Bool("feed", false, "Scrape recipe feed instead of a single recipe")

	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage of %s [options] [url]:\n", os.Args[0])
		flag.PrintDefaults()
		_, _ = fmt.Fprint(os.Stderr, "\nScrapes a Recipe data from a given webpage. Provide an URL to a valid HTML5 document.\n")
	}
	flag.Parse()

	switch len(flag.Args()) {
	case 1:
		targetUrl := flag.Args()[0]

		if *feedFlag {
			feed, err := krip.ScrapeFeedUrl(targetUrl)
			if err != nil {
				log.Fatal("Unable to scrape feed: " + err.Error())
			}
			fmt.Println(feed)
		} else {
			recipe, err := krip.ScrapeUrl(targetUrl)
			if err != nil {
				log.Fatal("Unable to scrape target: " + err.Error())
			}
			fmt.Println(recipe)
		}
	default:
		flag.Usage()
		os.Exit(1)
	}
}
