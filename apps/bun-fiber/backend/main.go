package main

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func main() {
	config := getConfig()
	initMetrics()
	go runMetricsServer()

	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		BodyLimit:    2 * 1024 * 1024,
	})

	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Format: "${time} | ${host} | ${method} | ${status} | ${latency} | ${ip} | ${path} | id=${locals:requestid} |\n",
	}))
	app.Use(safeHeadersMiddleware())
	app.Use(metricsMiddleware())

	// API group
	api := app.Group("/api", apiAuthMiddleware(config))
	app.Use(staticFilesMiddleware(config))

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
