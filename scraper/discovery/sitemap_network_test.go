package discovery

import (
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
			_, _ = w.Write([]byte("Sitemap: https://example.com/from-robots.xml\n"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	baseUrl, err := url.Parse(server.URL)
	require.NoError(t, err)
	data := &model.DataInput{Url: server.URL, RequestOptions: model.RequestOptions{HttpClient: server.Client()}}

	candidates := sitemapCandidates(data, baseUrl)

	require.Contains(t, candidates, "https://example.com/from-robots.xml")
	// robots.txt candidate must be tried before the hardcoded fallback guesses.
	robotsIdx := indexOf(candidates, "https://example.com/from-robots.xml")
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
	opts := model.RequestOptions{HttpClient: server.Client()}

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
	for i := 0; i < 6; i++ {
		indexLocs = append(indexLocs, fmt.Sprintf("%s/child-%d.xml", server.URL, i))
	}
	opts := model.RequestOptions{HttpClient: server.Client()}

	_, err := followSitemapIndex(indexLocs, opts)

	require.NoError(t, err)
	assert.Equal(t, int32(3), atomic.LoadInt32(&requests), "only the first 3 (bounded) children should be fetched")
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
