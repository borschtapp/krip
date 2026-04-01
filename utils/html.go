package utils

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
)

// GetMetaContent retrieves the "content" attribute of a meta tag matching the specified key and name in the selection.
func GetMetaContent(el *goquery.Selection, key, name string) (string, bool) {
	if val, ok := el.Find(fmt.Sprintf("meta[%s='%s']", key, name)).Attr("content"); ok {
		return val, true
	}

	return "", false
}

// FirstMetaContent returns the content attribute of the first matching meta-element from a list of identifiers.
func FirstMetaContent(el *goquery.Selection, names ...string) (string, bool) {
	for _, name := range names {
		if val, ok := GetMetaContent(el, "property", name); ok {
			return val, true
		}
		if val, ok := GetMetaContent(el, "name", name); ok {
			return val, true
		}
	}

	return "", false
}
