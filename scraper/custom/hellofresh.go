package custom

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/utils"
)

type nextData struct {
	Props struct {
		PageProps struct {
			SsrPayload struct {
				ActiveWeek string `json:"activeWeek"`
				Config     struct {
					HeadMeta struct {
						BrandName string `json:"brandName"`
					} `json:"head-meta-tags-feature-config"`
					Header struct {
						Logo struct {
							LogoURL string `json:"logoURL"`
						} `json:"logo"`
					} `json:"hf.funnel.header"`
				} `json:"config"`
				Courses []struct {
					Index  int `json:"index"`
					Recipe struct {
						Country  string `json:"country"`
						Cuisines []struct {
							ID       string `json:"id"`
							Type     string `json:"type"`
							Name     string `json:"name"`
							Slug     string `json:"slug"`
							IconLink string `json:"iconLink"`
						} `json:"cuisines"`
						Difficulty     int    `json:"difficulty"`
						FavoritesCount int    `json:"favoritesCount"`
						Headline       string `json:"headline"`
						ID             string `json:"id"`
						ImageLink      string `json:"imageLink"`
						ImagePath      string `json:"imagePath"`
						IsPublished    bool   `json:"isPublished"`
						Label          struct {
							Text   string `json:"text"`
							Handle string `json:"handle"`
						} `json:"label"`
						Name         string `json:"name"`
						PrepTime     string `json:"prepTime"`
						RatingsCount int    `json:"ratingsCount"`
						Slug         string `json:"slug"`
						Tags         []struct {
							ID   string `json:"id"`
							Type string `json:"type"`
							Name string `json:"name"`
							Slug string `json:"slug"`
						} `json:"tags"`
						TotalTime  string `json:"totalTime"`
						UUID       string `json:"uuid"`
						WebsiteURL string `json:"websiteUrl"`
					} `json:"recipe"`
				} `json:"courses"`
			} `json:"ssrPayload"`
		} `json:"pageProps"`
	} `json:"props"`
}

func ScrapeHelloFreshFeed(data *model.DataInput, feed *model.Feed) error {
	if data.Document == nil {
		return fmt.Errorf("no document found")
	}

	baseUrl, err := url.Parse(data.Url)
	if err != nil {
		return err
	}

	// Verify that the page is a menus page (e.g., /menus, /menus/, or regional variants)
	if !strings.HasPrefix(baseUrl.Path, "/menus") {
		return fmt.Errorf("not a menus page")
	}

	nextDataRaw := data.Document.Find("script#__NEXT_DATA__").Text()
	if nextDataRaw == "" {
		return fmt.Errorf("no next data found")
	}

	var nextDataObj nextData
	if err := json.Unmarshal([]byte(nextDataRaw), &nextDataObj); err != nil {
		return fmt.Errorf("json unmarshal error: %v", err)
	}

	payload := nextDataObj.Props.PageProps.SsrPayload

	var publisher *model.Organization
	if payload.Config.HeadMeta.BrandName != "" {
		publisher = &model.Organization{
			Name: payload.Config.HeadMeta.BrandName,
			Logo: payload.Config.Header.Logo.LogoURL,
			Url:  utils.BaseUrl(data.Url),
		}
	}

	var weekDate *time.Time
	if payload.ActiveWeek != "" {
		var year, week int
		if n, _ := fmt.Sscanf(payload.ActiveWeek, "%d-W%d", &year, &week); n == 2 {
			// Rough estimation of the week's start date
			t := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
			t = t.AddDate(0, 0, (week-1)*7)
			weekDate = &t
		}
	}

	uniqueEntries := make(map[string]*model.Recipe)

	for _, course := range payload.Courses {
		if course.Recipe.Name != "" {
			var categories []string
			var cuisines []string
			for _, t := range course.Recipe.Tags {
				categories = append(categories, t.Name)
			}
			for _, c := range course.Recipe.Cuisines {
				cuisines = append(cuisines, c.Name)
			}

			entry := &model.Recipe{
				Name:          course.Recipe.Name,
				Url:           course.Recipe.WebsiteURL,
				Description:   course.Recipe.Headline,
				Publisher:     publisher,
				TotalTime:     course.Recipe.TotalTime,
				Cuisines:      cuisines,
				Categories:    categories,
				Difficulty:    fmt.Sprint(course.Recipe.Difficulty),
				DatePublished: weekDate,
			}
			entry.AddImageUrl(course.Recipe.ImageLink)
			uniqueEntries[entry.Url] = entry
		}
	}

	for _, entry := range uniqueEntries {
		feed.Entries = append(feed.Entries, entry)
	}

	return nil
}
