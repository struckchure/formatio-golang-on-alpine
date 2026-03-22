package main

import (
	"time"
	"os"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"app":      "gofiber - demo",
			"datetime": time.Now().Format("2006-01-02 3:4:5 PM"),
			"email": os.Getenv("EMAIL")
		})
	})

	app.Listen(":8000")
}
