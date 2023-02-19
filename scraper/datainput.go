package scraper

import (
	"bytes"
	"errors"
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
		return nil, errors.New("unable to read the file: " + err.Error())
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, errors.New("failed to read the file: " + err.Error())
	}

	contentType := http.DetectContentType(content)
	if err == nil && strings.HasPrefix(contentType, "text/html") {
		root, err := html.Parse(bytes.NewReader(content))
		if err != nil {
			return nil, errors.New("unable to parse html tree: " + err.Error())
		}

		url := "file://" + strings.ReplaceAll(fileName, "\\", "/")
		return NodeInput(root, url, options)
	} else {
		return &model.DataInput{
			Text: string(content),
		}, nil
	}
}

func UrlInput(url string) (i *model.DataInput, err error) {
	resp, err := utils.ReadUrl(url, map[string][]string{
		"Accept":     {"text/html"},
		"User-Agent": {"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:101.0) Gecko/20100101 Firefox/101.0"},
	})
	if err != nil {
		return nil, err
	}

	root, err := html.Parse(bytes.NewReader(resp))
	if err != nil {
		return nil, errors.New("unable to parse html tree: " + err.Error())
	}

	input, err := NodeInput(root, url, model.InputOptions{})
	if err != nil {
		return nil, err
	}

	return input, nil
}

func NodeInput(root *html.Node, url string, options model.InputOptions) (i *model.DataInput, err error) {
	doc := goquery.NewDocumentFromNode(root)

	if !options.SkipUrl {
		if val, ok := doc.Find("link[rel='canonical']").Attr("href"); ok {
			url = val
		} else if val, ok := doc.Find("meta[property='og:url']").Attr("content"); ok {
			url = val
		} else if val, ok := doc.Find("link[rel='alternate']").Attr("href"); ok {
			url = val
		}
	}

	var schema *microdata.Item
	if !options.SkipSchema {
		data, err := microdata.ParseNode(root, url)
		if err != nil {
			log.Println("unable to parse microdata on the page: " + err.Error())
		} else {
			schema = data.GetFirstOfType("Recipe", "http://schema.org/Recipe", "https://schema.org/Recipe")
		}
	}

	return &model.DataInput{
		Url:      url,
		RootNode: root,
		Document: doc,
		Schema:   schema,
	}, nil
}
