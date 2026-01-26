package custom_test

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/borschtapp/krip"
	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/scraper/custom"
	"github.com/stretchr/testify/assert"
)

func TestMarleySpoonOnline(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping online test in short mode")
	}

	var website = "https://marleyspoon.de/menu/516754-vegetarisches-bibimbap-mit-spiegelei-und-kimchi"
	recipe, err := krip.ScrapeUrl(website)
	assert.NoError(t, err)
	assert.True(t, recipe.IsValid())
	t.Log(recipe.String())
}

func TestMarleySpoonFeedOnline(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping online test in short mode")
	}

	var website = "https://marleyspoon.de/menu"
	feed, err := krip.ScrapeFeedUrl(website, krip.FeedOptions{Quick: true})
	assert.NoError(t, err)
	assert.NotNil(t, feed)

	assert.NotEmpty(t, feed.Url)
	assert.NotEmpty(t, feed.Entries)
	for _, entry := range feed.Entries {
		assert.False(t, entry.IsEmpty())
	}
	t.Log(feed.String())
}

func TestMarleySpoonFeedParsing(t *testing.T) {
	content := `<main>
			<a href="/menu/552466-vegane-fish-n-chips-aus-tofu-mit-kartoffelecken-und-salat">
				<div
					class="flex shrink-0 flex-col justify-between rounded-container dn:bg-neutral-white dn:shadow-[0_2px_4px_0_rgba(0,0,0,0.08)] size-full">
					<div class="relative h-[180px] cursor-pointer">
						<div class="absolute right-2 top-2 z-10"></div><img
							class="absolute inset-0 size-full select-none object-cover object-center rounded-t-container"
							alt="Vegane Fish ’n’ Chips aus Tofu" loading="lazy"
							src="https://marleyspoon.com/media/recipes/552466/main_photos/medium/fish_n_chips_vegan_aus_tofu-4378114a822c5b014e2182e1b6577cdf.jpeg">
					</div>
					<div class="flex grow flex-col p-4">
						<div class="mb-4 cursor-pointer">
							<div class="font-sans text-sm font-strong leading-sm">Vegane Fish ’n’ Chips aus Tofu</div>
							<div class="font-sans text-xs leading-xs mt-1">mit Kartoffelecken und Salat</div>
						</div>
						<div class="flex flex-col gap-2">
							<div class="flex gap-2">
								<div class="flex items-center gap-1"><img class="size-4" alt="Allergen"
										src="https://marleyspoon.com/media/findability/ms/no_dairy.png"></div>
								<div class="font-sans text-xs font-strong leading-xs">20-30 Minuten</div>
							</div>
							<div class="flex flex-wrap gap-0.5 text-neutral-greyDark dn:text-neutral-black">
								<div class="font-sans text-xs leading-xs">
									<div class="flex items-start justify-between gap-1">
										<div
											class="font-sans text-xs font-strong leading-xs grow whitespace-nowrap lowercase text-tertiary-ultraDark first-letter:capitalize dn:text-primary-veryDark">
											Vegetarisch</div>
									</div>
								</div>
								<div class="font-sans text-xs leading-xs mx-px">•</div>
								<div class="font-sans text-xs leading-xs">Familienfreundlich</div>
								<div class="font-sans text-xs leading-xs mx-px">•</div>
								<div class="font-sans text-xs leading-xs">Good climate score</div>
								<div class="font-sans text-xs leading-xs mx-px">•</div>
								<div class="font-sans text-xs leading-xs">Protein+</div>
							</div>
						</div>
					</div>
				</div>
			</a>
		</main>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(content)))
	if err != nil {
		t.Fatal(err)
	}

	data := &model.DataInput{
		Url:      "https://marleyspoon.de/menu",
		Document: doc,
	}

	feed := &model.Feed{Url: data.Url}
	assert.NoError(t, custom.ScrapeMarleySpoonFeed(data, feed))
	assert.Len(t, feed.Entries, 1)

	entry := feed.Entries[0]
	assert.Equal(t, "Vegane Fish ’n’ Chips aus Tofu mit Kartoffelecken und Salat", entry.Name)
	assert.Equal(t, "PT30M", entry.TotalTime)
	assert.False(t, len(entry.Images) == 0)
	assert.Equal(t, "https://marleyspoon.com/media/recipes/552466/main_photos/medium/fish_n_chips_vegan_aus_tofu-4378114a822c5b014e2182e1b6577cdf.jpeg", entry.Images[0].Url)
	assert.ElementsMatch(t, entry.Categories, []string{"Vegetarisch", "Familienfreundlich", "Good climate score", "Protein+"})
	t.Log(feed.String())
}
