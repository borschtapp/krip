package scraper

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"

	"github.com/astappiev/microdata"
	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/utils"
)

func FileInput(fileName string, options model.InputOptions) (*model.DataInput, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("unable to read the file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read the file: %w", err)
	}

	contentType := http.DetectContentType(content)
	if strings.HasPrefix(contentType, "text/html") {
		root, err := html.Parse(bytes.NewReader(content))
		if err != nil {
			return nil, fmt.Errorf("unable to parse html tree: %w", err)
		}

		url := "file://" + strings.ReplaceAll(fileName, "\\", "/")
		return NodeInput(root, url, options)
	}

	return &model.DataInput{
		Text: string(content),
	}, nil
}

func UrlInput(url string) (*model.DataInput, error) {
	resp, respUrl, err := utils.ReadUrl(url, map[string][]string{
		"Accept":     {"text/html"},
		"Referer":    {"https://www.google.com/"},
		"User-Agent": {"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/110.0"},
	})
	if err != nil {
		return nil, err
	}

	root, err := html.Parse(bytes.NewReader(resp))
	if err != nil {
		return nil, fmt.Errorf("unable to parse html tree: %w", err)
	}

	input, err := NodeInput(root, respUrl.String(), model.InputOptions{SkipUrl: true})
	if err != nil {
		return nil, err
	}

	return input, nil
}

func NodeInput(root *html.Node, url string, options model.InputOptions) (*model.DataInput, error) {
	doc := goquery.NewDocumentFromNode(root)

	if !options.SkipUrl { // if we read the page from a file, we need to retrieve a url
		if val, ok := doc.Find("link[rel='canonical']").Attr("href"); ok && utils.IsAbsolute(val) {
			url = val
		} else if val, ok := doc.Find("meta[property='og:url']").Attr("content"); ok && utils.IsAbsolute(val) {
			url = val
		} else if val, ok := doc.Find("link[rel='alternate']").Attr("href"); ok && utils.IsAbsolute(val) {
			url = val
		}
	}

	var err error
	var schemas *microdata.Microdata
	if !options.SkipSchema {
		schemas, err = microdata.ParseNode(root, url)
		if err != nil {
			log.Println("unable to parse microdata on the page: " + err.Error())
		}
	}

	return &model.DataInput{
		Url:      url,
		RootNode: root,
		Document: doc,
		Schemas:  schemas,
	}, nil
}
