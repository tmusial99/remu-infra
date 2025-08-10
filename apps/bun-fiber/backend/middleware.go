package main

import "github.com/gofiber/fiber/v2"

func apiAuthMiddleware(config *ServerConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		host := cleanHost(c.Get("Host"))

		if host == "localhost" || host == "remu" {
			host = config.LocalhostTestDomain
		}

		for _, domain := range config.Domains {
			if domain.Host == host {
				if domain.APIEnabled {
					return c.Next()
				}
				break
			}
		}

		return c.Status(404).SendString("API endpoint not found for this domain")
	}
}