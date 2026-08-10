package scraper

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/astappiev/microdata"
	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/utils"
	"golang.org/x/net/html"
)

func FileInput(fileName string, options model.ScrapeOptions) (*model.DataInput, error) {
	file, err := os.Open(filepath.Clean(fileName))
	if err != nil {
		return nil, fmt.Errorf("unable to read the file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

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
		input, err := NodeInput(root, url, options)
		if err != nil {
			return nil, err
		}

		input.Text = string(content)
		input.RequestOptions = options.RequestOptions
		return input, nil
	}

	return &model.DataInput{
		Text:           string(content),
		RequestOptions: options.RequestOptions,
	}, nil
}

func UrlInput(url string, options model.ScrapeOptions) (*model.DataInput, error) {
	resp, err := utils.ExecuteRequest(utils.RequestConfig{URL: url}, options.RequestOptions)
	if err != nil {
		return nil, err
	}

	contentType := resp.ContentType
	if contentType == "" {
		contentType = http.DetectContentType(resp.Body)
	}
	if !strings.HasPrefix(contentType, "text/html") && !strings.HasPrefix(contentType, "application/xhtml+xml") {
		return &model.DataInput{
			Url:            resp.URL.String(),
			Text:           string(resp.Body),
			RequestOptions: options.RequestOptions,
		}, nil
	}

	root, err := html.Parse(bytes.NewReader(resp.Body))
	if err != nil {
		return nil, fmt.Errorf("unable to parse html tree: %w", err)
	}

	nodeOpts := options
	nodeOpts.SkipMetaUrl = true
	input, err := NodeInput(root, resp.URL.String(), nodeOpts)
	if err != nil {
		return nil, err
	}

	input.Text = string(resp.Body)
	input.RequestOptions = options.RequestOptions
	return input, nil
}

func NodeInput(root *html.Node, url string, options model.ScrapeOptions) (*model.DataInput, error) {
	doc := goquery.NewDocumentFromNode(root)

	if !options.SkipMetaUrl { // if we read the page from a file, we need to retrieve a url
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
	if !options.SkipMicrodata {
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
