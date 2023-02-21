package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/borschtapp/krip"
)

//go:embed static/*
var static embed.FS

func scrapeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	q := r.URL.Query()

	if q == nil || len(q.Get("url")) == 0 {
		http.Error(w, "`url` query param is required.", http.StatusBadRequest)
		return
	}

	recipe, err := krip.ScrapeUrl(q.Get("url"))
	if err != nil {
		http.Error(w, "Scrape error: "+err.Error(), http.StatusInternalServerError)
	}

	j, _ := json.Marshal(recipe)
	_, _ = w.Write(j)
}

func main() {
	subFs, err := fs.Sub(static, "static")
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", http.FileServer(http.FS(subFs)))
	http.HandleFunc("/api/v1/scrape", scrapeHandler)

	fmt.Printf("Starting server at port http://localhost:3000\n")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatal(err)
	}
}
