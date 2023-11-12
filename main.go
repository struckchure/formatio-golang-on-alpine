package main

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"app":      "gofiber",
			"datetime": time.Now().Format("2006-01-02 3:4:5 PM"),
		})
	})

	app.Listen(":8000")
}
