package discovery

import (
	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/utils"
)

// RefineByUrlPattern derives the common URL path prefix from validEntries and returns
// the pattern and count of allEntries whose URL matches that prefix.
func RefineByUrlPattern(allEntries, validEntries []*model.Recipe) (string, int) {
	urls := make([]string, 0, len(validEntries))
	for _, e := range validEntries {
		if e.Url != "" {
			urls = append(urls, e.Url)
		}
	}
	pattern := UrlPathPattern(urls)
	if pattern == "" {
		return "", 0
	}
	count := 0
	for _, e := range allEntries {
		if utils.UrlMatchesPathPattern(e.Url, pattern) {
			count++
		}
	}
	return pattern, count
}
