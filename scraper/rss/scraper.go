package rss

import (
	"fmt"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/utils"
	"github.com/mmcdole/gofeed"
)

func ScrapeFeed(data *model.DataInput, feed *model.Feed) error {
	if data.Text == "" {
		return fmt.Errorf("no text")
	}

	fp := gofeed.NewParser()
	f, err := fp.ParseString(data.Text)
	if err != nil {
		return fmt.Errorf("rss parsing error: %w", err)
	}

	var publisher *model.Organization
	if f.Title != "" || f.Description != "" {
		publisher = &model.Organization{
			Name:        f.Title,
			Description: f.Description,
			Url:         f.Link,
		}
		if f.Image != nil {
			publisher.Logo = f.Image.URL
		}
	}

	for _, item := range f.Items {
		entry := &model.Recipe{
			Url:           item.Link,
			Name:          item.Title,
			Publisher:     publisher,
			Description:   utils.Cleanup(item.Description),
			Categories:    item.Categories,
			DatePublished: item.PublishedParsed,
		}

		if item.Image != nil {
			entry.AddImageUrl(item.Image.URL)
		} else if len(item.Extensions["media"]["content"]) > 0 {
			entry.AddImageUrl(item.Extensions["media"]["content"][0].Attrs["url"])
		}

		if len(item.Authors) > 0 {
			entry.Author = &model.Person{
				Name: item.Authors[0].Name,
			}
		}

		feed.Entries = append(feed.Entries, entry)
	}

	return nil
}
