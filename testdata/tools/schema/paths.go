package main

import (
	"fmt"
	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/scraper"
	"github.com/borschtapp/krip/testdata"
	"log"
	"os"
	"sort"
	"strings"
)

func main() {
	var paths = make(map[string]int)

	testdata.WalkTestdataWebsites(func(name string, path string) {
		input, err := scraper.FileInput(path, model.InputOptions{SkipText: true})
		if err != nil {
			log.Fatal(err)
		}

		if input.Schemas != nil {
			input.Schemas.GetFirstOfSchemaType("Recipe").CountPaths("", &paths)
		}
	})

	lines := make([]string, 0, len(paths))
	for key, count := range paths {
		lines = append(lines, fmt.Sprint(key, " ", count))
	}
	sort.Strings(lines)
	text := strings.Join(lines, "\n")

	_ = os.WriteFile(testdata.PackageDir+"schema_paths.txt", []byte(text), 0644)
}
