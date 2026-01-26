package rss

import (
	"testing"

	"github.com/borschtapp/krip/model"
	"github.com/stretchr/testify/assert"
)

func TestScrapeRSS(t *testing.T) {
	rssText := `<?xml version="1.0" encoding="UTF-8" ?>
<rss version="2.0" xmlns:media="http://search.yahoo.com/mrss/">
<channel>
 <title>RSS Title</title>
 <description>This is an example of an RSS feed</description>
 <link>http://www.example.com/main.html</link>
 <image>
  <url>http://www.example.com/logo.png</url>
 </image>
 <item>
  <title>Example Entry</title>
  <description>Here is some text in the description.</description>
  <link>http://www.example.com/blog/post/1</link>
  <category>Recipes</category>
  <category>Dinner</category>
  <pubDate>Sun, 25 Jan 2026 10:00:00 GMT</pubDate>
  <dc:creator xmlns:dc="http://purl.org/dc/elements/1.1/">John Doe</dc:creator>
  <media:content url="http://www.example.com/image.jpg" medium="image" />
 </item>
</channel>
</rss>`

	data := &model.DataInput{
		Url:  "http://www.example.com/rss",
		Text: rssText,
	}

	feed := &model.Feed{Url: data.Url}
	assert.NoError(t, ScrapeFeed(data, feed))
	assert.Equal(t, 1, len(feed.Entries))
	entry := feed.Entries[0]
	assert.NotNil(t, entry.Publisher)
	assert.Equal(t, "RSS Title", entry.Publisher.Name)
	assert.Equal(t, "http://www.example.com/logo.png", entry.Publisher.Logo)

	assert.Equal(t, "Example Entry", entry.Name)
	assert.Equal(t, "http://www.example.com/blog/post/1", entry.Url)
	assert.Len(t, entry.Images, 1)
	assert.Equal(t, "http://www.example.com/image.jpg", entry.Images[0].Url)
	assert.Equal(t, []string{"Recipes", "Dinner"}, entry.Categories)
	assert.NotNil(t, entry.Author)
	assert.Equal(t, "John Doe", entry.Author.Name)
	assert.NotNil(t, entry.DatePublished)
	t.Log(feed.String())
}

func TestScrapeAtom(t *testing.T) {
	atomText := `<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
 <title>Example Feed</title>
 <logo>http://example.org/logo.png</logo>
 <entry>
  <title>Atom Entry</title>
  <link href="http://example.org/2003/12/13/atom03" />
  <summary>Some summary.</summary>
  <category term="Lunch" />
  <published>2026-01-25T11:00:00Z</published>
  <author>
   <name>Jane Smith</name>
  </author>
 </entry>
</feed>`

	data := &model.DataInput{
		Url:  "http://example.org/feed",
		Text: atomText,
	}

	feed := &model.Feed{Url: data.Url}
	assert.NoError(t, ScrapeFeed(data, feed))
	assert.Equal(t, 1, len(feed.Entries))
	entry := feed.Entries[0]
	assert.NotNil(t, entry.Publisher)
	assert.Equal(t, "http://example.org/logo.png", entry.Publisher.Logo)
	assert.Equal(t, "Atom Entry", entry.Name)
	assert.Equal(t, "http://example.org/2003/12/13/atom03", entry.Url)
	assert.Equal(t, []string{"Lunch"}, entry.Categories)
	assert.NotNil(t, entry.Author)
	assert.Equal(t, "Jane Smith", entry.Author.Name)
	assert.NotNil(t, entry.DatePublished)
	t.Log(feed.String())
}
