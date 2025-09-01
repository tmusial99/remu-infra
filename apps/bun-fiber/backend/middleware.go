package main

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func apiAuthMiddleware(config *ServerConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		host := cleanHost(c.Hostname())

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

func safeHeadersMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// edit only these lists
		scriptHosts := []string{
			"https://maps.googleapis.com",
			"https://maps.gstatic.com",
			"https://app.reviewconnect.me",
			"https://static.cloudflareinsights.com",
		}
		styleHosts := []string{
			"https://fonts.googleapis.com",
			"https://app.reviewconnect.me",
		}
		fontHosts := []string{
			"https://fonts.gstatic.com",
		}
		connectHosts := []string{
			"https://maps.googleapis.com",
			"https://maps.gstatic.com",
			"https://app.reviewconnect.me",
		}
		frameHosts := []string{
			"https://www.google.com/maps/",
		}

		pad := func(hosts []string) string {
			if len(hosts) == 0 {
				return ""
			}
			return " " + strings.Join(hosts, " ")
		}

		// security headers
		c.Set(fiber.HeaderXContentTypeOptions, "nosniff")
		c.Set(fiber.HeaderReferrerPolicy, "strict-origin-when-cross-origin")
		c.Set(fiber.HeaderXFrameOptions, "DENY")
		c.Set(fiber.HeaderPermissionsPolicy, "geolocation=(), microphone=(), camera=(), payment=(), usb=(), interest-cohort=()")

		// CSP – single-line header
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline'" + pad(scriptHosts) + "; " +
			"style-src 'self' 'unsafe-inline'" + pad(styleHosts) + "; " +
			"img-src 'self' data: https:; " +
			"font-src 'self' data:" + pad(fontHosts) + "; " +
			"connect-src 'self'" + pad(connectHosts) + "; " +
			"frame-src 'self'" + pad(frameHosts) + "; " +
			"object-src 'none'; " +
			"base-uri 'self'; " +
			"frame-ancestors 'none'; "

		c.Set(fiber.HeaderContentSecurityPolicy, csp)
		return c.Next()
	}
}
