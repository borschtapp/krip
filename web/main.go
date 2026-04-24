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

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	j, err := json.Marshal(v)
	if err != nil {
		http.Error(w, "JSON encoding error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(j)
}

func requireURL(w http.ResponseWriter, r *http.Request) (string, bool) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return "", false
	}
	u := r.URL.Query().Get("url")
	if u == "" {
		http.Error(w, "`url` query param is required.", http.StatusBadRequest)
		return "", false
	}
	return u, true
}

func scrapeHandler(w http.ResponseWriter, r *http.Request) {
	u, ok := requireURL(w, r)
	if !ok {
		return
	}
	recipe, err := krip.ScrapeUrl(u, krip.ScrapeOptions{})
	if err != nil {
		http.Error(w, "Scrape error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, recipe)
}

func feedHandler(w http.ResponseWriter, r *http.Request) {
	u, ok := requireURL(w, r)
	if !ok {
		return
	}
	feed, err := krip.ScrapeFeedUrl(u, krip.FeedOptions{})
	if err != nil {
		http.Error(w, "Scrape error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, feed)
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func main() {
	subFs, err := fs.Sub(static, "static")
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", http.FileServer(http.FS(subFs)))
	http.HandleFunc("/_health", healthHandler)
	http.HandleFunc("/api/v1/scrape", scrapeHandler)
	http.HandleFunc("/api/v1/feed", feedHandler)

	fmt.Printf("Starting server at port http://localhost:3000\n")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatal(err)
	}
}
