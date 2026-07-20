package discovery

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/utils"
)

func TestRobotsSitemaps_ParsesDirective(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/robots.txt" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte("User-agent: *\nDisallow: /admin\nSitemap: https://example.com/custom-sitemap.xml # with trailing comment\nsitemap: https://example.com/other-sitemap.xml\n# Sitemap: https://example.com/commented-out.xml\n"))
	}))
	defer server.Close()

	sitemaps := robotsSitemaps(server.URL, model.RequestOptions{HttpClient: server.Client()})
	assert.Equal(t, []string{"https://example.com/custom-sitemap.xml", "https://example.com/other-sitemap.xml"}, sitemaps,
		"should parse Sitemap: directives case-insensitively and strip comments")
}

func TestRobotsSitemaps_NoRobotsTxt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	sitemaps := robotsSitemaps(server.URL, model.RequestOptions{HttpClient: server.Client()})
	assert.Nil(t, sitemaps)
}

func TestSitemapCandidates_IncludesRobotsTxtSitemap(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			fmt.Fprintf(w, "Sitemap: http://%s/from-robots.xml\n", r.Host)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	baseUrl, err := url.Parse(server.URL)
	require.NoError(t, err)
	data := &model.DataInput{Url: server.URL, RequestOptions: model.RequestOptions{HttpClient: server.Client()}}

	candidates := sitemapCandidates(data, baseUrl)

	expectedSitemap := server.URL + "/from-robots.xml"
	require.Contains(t, candidates, expectedSitemap)
	// robots.txt candidate must be tried before the hardcoded fallback guesses.
	robotsIdx := indexOf(candidates, expectedSitemap)
	fallbackIdx := indexOf(candidates, server.URL+"/sitemap_index.xml")
	assert.Less(t, robotsIdx, fallbackIdx, "robots.txt sitemap should be probed before hardcoded guesses")
}

func TestSitemapCandidates_Deduplicates(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			fmt.Fprintf(w, "Sitemap: %s/sitemap.xml\n", server.URL)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	baseUrl, err := url.Parse(server.URL)
	require.NoError(t, err)
	data := &model.DataInput{Url: server.URL, RequestOptions: model.RequestOptions{HttpClient: server.Client()}}

	candidates := sitemapCandidates(data, baseUrl)

	count := 0
	for _, c := range candidates {
		if c == server.URL+"/sitemap.xml" {
			count++
		}
	}
	assert.Equal(t, 1, count, "duplicate sitemap candidate URLs should be filtered out")
}

func indexOf(s []string, v string) int {
	for i, x := range s {
		if x == v {
			return i
		}
	}
	return -1
}

// TestFollowSitemapIndex_MergesConcurrently verifies that fetching child sitemaps
// concurrently still produces the same de-duplicated result as sequential fetching,
// and that all (and only) the selected children are requested.
func TestFollowSitemapIndex_MergesConcurrently(t *testing.T) {
	var requests int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		switch r.URL.Path {
		case "/child-a.xml":
			_, _ = w.Write(urlsetXML("https://example.com/recipes/a1", "https://example.com/recipes/a2"))
		case "/child-b.xml":
			_, _ = w.Write(urlsetXML("https://example.com/recipes/b1", "https://example.com/recipes/a1")) // a1 overlaps child-a
		case "/child-c.xml":
			_, _ = w.Write(urlsetXML("https://example.com/recipes/c1"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	indexLocs := []string{server.URL + "/child-a.xml", server.URL + "/child-b.xml", server.URL + "/child-c.xml"}
	opts := model.RequestOptions{
		HttpClient: server.Client(),
		Context:    context.WithValue(context.Background(), utils.AllowLoopbackKey, true),
	}

	locs, err := followSitemapIndex(indexLocs, opts)

	require.NoError(t, err)
	assert.Equal(t, int32(3), atomic.LoadInt32(&requests), "all 3 selected children should be fetched")
	assert.ElementsMatch(t, []string{
		"https://example.com/recipes/a1",
		"https://example.com/recipes/a2",
		"https://example.com/recipes/b1",
		"https://example.com/recipes/c1",
	}, locs, "duplicate loc across children must be merged")
}

func TestFollowSitemapIndex_CapsAtThreeChildren(t *testing.T) {
	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		n := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/child-"), ".xml")
		_, _ = w.Write(urlsetXML(fmt.Sprintf("https://example.com/recipes/%s", n)))
	}))
	defer server.Close()

	var indexLocs []string
	for i := range 6 {
		indexLocs = append(indexLocs, fmt.Sprintf("%s/child-%d.xml", server.URL, i))
	}
	opts := model.RequestOptions{
		HttpClient: server.Client(),
		Context:    context.WithValue(context.Background(), utils.AllowLoopbackKey, true),
	}

	_, err := followSitemapIndex(indexLocs, opts)

	require.NoError(t, err)
	assert.Equal(t, int32(3), atomic.LoadInt32(&requests), "only the first 3 (bounded) children should be fetched")
}

func TestFollowSitemapIndex_NestedIndexRecursion(t *testing.T) {
	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		switch r.URL.Path {
		case "/nested-index.xml":
			// returns a sitemap index pointing to leaf-a and leaf-b
			_, _ = w.Write([]byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <sitemap><loc>http://%s/leaf-a.xml</loc></sitemap>
  <sitemap><loc>http://%s/leaf-b.xml</loc></sitemap>
</sitemapindex>`, r.Host, r.Host)))
		case "/leaf-a.xml":
			_, _ = w.Write(urlsetXML("https://example.com/recipes/a1"))
		case "/leaf-b.xml":
			_, _ = w.Write(urlsetXML("https://example.com/recipes/b1"))
		case "/leaf-c.xml":
			_, _ = w.Write(urlsetXML("https://example.com/recipes/c1"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// indexLocs: first one is an index, second is a leaf
	indexLocs := []string{server.URL + "/nested-index.xml", server.URL + "/leaf-c.xml"}
	opts := model.RequestOptions{
		HttpClient: server.Client(),
		Context:    context.WithValue(context.Background(), utils.AllowLoopbackKey, true),
	}

	locs, err := followSitemapIndex(indexLocs, opts)

	require.NoError(t, err)
	// Fetches:
	// 1. /nested-index.xml (index)
	// 2. /leaf-a.xml (leaf inside nested-index)
	// 3. /leaf-b.xml (leaf inside nested-index)
	// Budget is now 0. /leaf-c.xml is NOT fetched.
	assert.Equal(t, int32(3), atomic.LoadInt32(&requests), "should perform exactly 3 fetches total")
	assert.ElementsMatch(t, []string{
		"https://example.com/recipes/a1",
		"https://example.com/recipes/b1",
	}, locs, "should retrieve locs from the followed nested index leaf sitemaps")
}

func TestFollowSitemapIndex_NestedIndexRecursionLimit(t *testing.T) {
	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		switch r.URL.Path {
		case "/nested-index-1.xml":
			_, _ = w.Write([]byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <sitemap><loc>http://%s/nested-index-2.xml</loc></sitemap>
</sitemapindex>`, r.Host)))
		case "/nested-index-2.xml":
			_, _ = w.Write([]byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <sitemap><loc>http://%s/leaf.xml</loc></sitemap>
</sitemapindex>`, r.Host)))
		case "/leaf.xml":
			_, _ = w.Write(urlsetXML("https://example.com/recipes/leaf"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	indexLocs := []string{server.URL + "/nested-index-1.xml"}
	opts := model.RequestOptions{
		HttpClient: server.Client(),
		Context:    context.WithValue(context.Background(), utils.AllowLoopbackKey, true),
	}

	locs, err := followSitemapIndex(indexLocs, opts)

	require.NoError(t, err)
	// Fetches:
	// 1. /nested-index-1.xml (returns index, depth 0 -> depth 1)
	// 2. /nested-index-2.xml (returns index, depth 1 -> depth 2, but depth limit is 1, so we do not recurse)
	// Total fetches should be 2, and no leaf locations should be found because the recursion stopped at the second index.
	assert.Equal(t, int32(2), atomic.LoadInt32(&requests))
	assert.Empty(t, locs)
}

func urlsetXML(locs ...string) []byte {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
	for _, loc := range locs {
		fmt.Fprintf(&b, "<url><loc>%s</loc></url>", loc)
	}
	b.WriteString("</urlset>")
	return []byte(b.String())
}
