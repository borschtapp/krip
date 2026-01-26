package model

import "encoding/json"

// Feed represents a list of recipes found on a page or in a feed
type Feed struct {
	Url     string    `json:"url"`
	Entries []*Recipe `json:"entries"`
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
