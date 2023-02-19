package main

import (
	"github.com/gofiber/fiber/v2"

	"github.com/borschtapp/krip"
)

type request struct {
	URL string `json:"url"`
}

func ScrapeURL(c *fiber.Ctx) error {
	// check for the incoming request body
	body := new(request)
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "cannot parse request",
			"message": err.Error(),
		})
	}

	recipe, err := krip.ScrapeUrl(body.URL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "unable to scrape target",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(recipe)
}
