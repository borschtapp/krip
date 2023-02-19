package main

import (
	"embed"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

//go:embed static/*
var static embed.FS

func main() {
	app := fiber.New()
	app.Use(recover.New())
	app.Use(logger.New())

	app.Use("/", filesystem.New(filesystem.Config{
		Root:       http.FS(static),
		PathPrefix: "static",
		Index:      "index.html",
	}))

	app.Post("/api/v1/scrape", ScrapeURL)

	log.Fatal(app.Listen(":3000"))
}
