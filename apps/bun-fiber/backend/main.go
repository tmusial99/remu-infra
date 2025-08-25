package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	config := getConfig()
	initMetrics()

	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
	})

	// Logging middleware
	app.Use(logger.New())
	app.Use(metricsMiddleware())
	app.Use(staticFilesMiddleware(config))
	go runMetricsServer()

	// API group
	api := app.Group("/api", apiAuthMiddleware(config))

	// API endpoints
	api.Get("/hello", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello World!",
			"domain":  c.Get("Host"),
			"path":    c.Path(),
		})
	})

	api.Get("/users", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"users":  []string{"John", "Jane", "Bob"},
			"domain": c.Get("Host"),
		})
	})

	api.Post("/data", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Data received",
			"body":    string(c.Body()),
		})
	})

	// Fallback 404
	app.Use("*", func(c *fiber.Ctx) error {
		return c.Status(404).SendString("Not Found")
	})

	printConfig(config)
	app.Listen(config.Port)
}
