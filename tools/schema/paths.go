package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/scraper"
	"github.com/borschtapp/krip/testdata"
)

func main() {
	var paths = make(map[string]int)

	_ = filepath.Walk(testdata.WebsitesDir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(info.Name(), testdata.HtmlExt) {
			input, err := scraper.FileInput(path, model.InputOptions{SkipText: true})
			if err != nil {
				log.Fatal(err)
			}

			if input.Schema != nil {
				input.Schema.CountPaths("", &paths)
			}
		}
		return nil
	})

	lines := make([]string, 0, len(paths))
	for key, count := range paths {
		lines = append(lines, fmt.Sprint(key, " ", count))
	}
	sort.Strings(lines)
	text := strings.Join(lines, "\n")

	_ = os.WriteFile(testdata.PackageDir+"schema_paths.txt", []byte(text), 0644)
}
