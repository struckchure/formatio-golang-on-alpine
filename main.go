package main

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	app := fiber.New()
	app.Use(logger.New(logger.Config{
		Format: "${time} ${status} - ${latency} ${method} ${path}\n",
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"app":      "gofiber - demo 3",
			"datetime": time.Now().Format("2006-01-02 3:4:5 PM"),
			"email":    os.Getenv("EMAIL"),
		})
	})

	app.Listen(":8000")
}
