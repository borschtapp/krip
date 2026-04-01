package model

import "encoding/json"

// DiscoverySource identifies the method used to discover a feed.
type DiscoverySource string

// DiscoveredFeed holds the result of automatic feed discovery.
// Serialize and pass back via FeedOptions.Discovered to skip re-discovery.
type DiscoveredFeed struct {
	// Source describes how the feed was discovered.
	Source DiscoverySource `json:"source"`
	// Selector holds the source-specific locator:
	// - "rss-link": the absolute RSS/Atom feed URL
	// - "sitemap": the sitemap URL that was fetched
	// - "dom-container": the structural CSS-path key of the container containing recipe links
	Selector string `json:"selector,omitempty"`
	// UrlPattern is the common path prefix shared by discovered recipe URLs, e.g. "/recipes/".
	UrlPattern string `json:"urlPattern,omitempty"`
	// ConfidenceScore is the scoring result for dom-container source (0.0–1.0).
	ConfidenceScore float64 `json:"confidenceScore,omitempty"`
}

// Feed represents a list of recipes found on a page or in a feed
type Feed struct {
	Name        string         `json:"name,omitempty"`
	Url         string         `json:"url,omitempty"`
	Description string         `json:"description,omitempty"`
	Language    string         `json:"inLanguage,omitempty"`
	Images      []*ImageObject `json:"image,omitempty"`
	Publisher   *Organization  `json:"publisher,omitempty"`
	Entries     []*Recipe      `json:"entries,omitempty"`
	// Discovered is populated when the universal discovery strategy was used.
	// Persist and pass back via FeedOptions.Discovered for faster subsequent calls.
	Discovered *DiscoveredFeed `json:"discovered,omitempty"`
}

// AddEntry adds a recipe to the feed if it does not already exist
func (f *Feed) AddEntry(entry *Recipe) bool {
	for _, e := range f.Entries {
		if (len(entry.Url) > 0 && e.Url == entry.Url) || (len(entry.Name) > 0 && e.Name == entry.Name) {
			return false
		}
	}

	f.Entries = append(f.Entries, entry)
	return true
}

func (f *Feed) String() string {
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return "Unable to output in json: " + err.Error()
	}
	return string(data)
}
