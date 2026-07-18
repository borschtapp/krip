package discovery

import (
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"

	"github.com/borschtapp/krip/model"
)

// genListingHTML builds a synthetic nav-heavy recipe listing page with n recipe
// cards, plus nav/sidebar/footer noise, mirroring the "500-3000 links" profile
// described in discovery-ideas.md. Cards are wrapped in extra single-child divs
// (like real card/grid layouts) so containerKey/effectiveParent/mergeSiblingGroups
// all do real work, not just a flat link list.
func genListingHTML(n int) string {
	var b strings.Builder
	b.WriteString("<!DOCTYPE html><html><body>")

	b.WriteString("<nav>")
	for i := 0; i < 40; i++ {
		fmt.Fprintf(&b, `<a href="/category/%d">Category %d</a>`, i, i)
	}
	b.WriteString("</nav>")

	b.WriteString(`<aside class="sidebar">`)
	for i := 0; i < 25; i++ {
		fmt.Fprintf(&b, `<a href="/popular/%d">Popular Post %d</a>`, i, i)
	}
	b.WriteString("</aside>")

	b.WriteString(`<main><section class="grid">`)
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, `<div class="tile"><div class="tile-inner"><a href="/recipes/recipe-%d"><img src="/img/recipe-%d.jpg"><span>Delicious Recipe Number %d</span></a></div></div>`, i, i, i)
	}
	b.WriteString("</section></main>")

	b.WriteString("<footer>")
	for i := 0; i < 20; i++ {
		fmt.Fprintf(&b, `<a href="/legal/%d">Legal Page %d</a>`, i, i)
	}
	b.WriteString("</footer>")

	b.WriteString("</body></html>")
	return b.String()
}

func parseDocumentB(b *testing.B, htmlContent string) *goquery.Document {
	b.Helper()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		b.Fatal(err)
	}
	return doc
}

func BenchmarkCollectGroups(b *testing.B) {
	for _, n := range []int{500, 1500, 3000} {
		b.Run(fmt.Sprintf("links=%d", n), func(b *testing.B) {
			doc := parseDocumentB(b, genListingHTML(n))
			base, _ := url.Parse("https://example.com/recipes")

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				collectGroups(doc, base)
			}
		})
	}
}

func BenchmarkTryDOMScoring(b *testing.B) {
	for _, n := range []int{500, 1500, 3000} {
		b.Run(fmt.Sprintf("links=%d", n), func(b *testing.B) {
			htmlContent := genListingHTML(n)
			base, _ := url.Parse("https://example.com/recipes")

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				doc := parseDocumentB(b, htmlContent)
				data := &model.DataInput{Url: "https://example.com/recipes", Document: doc}
				feed := &model.Feed{}
				b.StartTimer()

				_, _ = tryDOMScoring(data, feed, base, SamplingOptions{}, ScoringOptions{})
			}
		})
	}
}
