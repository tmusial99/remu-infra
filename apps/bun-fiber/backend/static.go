package main

import (
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type cachedFile struct {
	data        []byte
	contentType string
}

func preloadStaticFiles(config *ServerConfig) map[string]cachedFile {
	cache := make(map[string]cachedFile)

	for _, domain := range config.Domains {
		filepath.Walk(domain.StaticPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}

			relPath, _ := filepath.Rel(domain.StaticPath, path)
			relPath = "/" + filepath.ToSlash(relPath)

			ext := filepath.Ext(path)
			ct := mime.TypeByExtension(ext)
			if ct == "" {
				ct = "application/octet-stream"
			}

			data, err := os.ReadFile(path)
			if err != nil {
				log.Printf("⚠️ Error reading %s: %v", path, err)
				return nil
			}

			cache[domain.Host+relPath] = cachedFile{
				data:        data,
				contentType: ct,
			}
			return nil
		})
	}

	return cache
}

func setCacheHeaders(c *fiber.Ctx, path string, config *ServerConfig) {
    if hasExtension(path, config.NoCacheExtensions) {
        c.Set(fiber.HeaderCacheControl, "public, max-age=300")
    } else {
        c.Set(fiber.HeaderCacheControl, "public, max-age=31536000, immutable")
    }
}

func staticFilesMiddleware(config *ServerConfig) fiber.Handler {
	cache := preloadStaticFiles(config)

	domainMap := make(map[string]DomainConfig, len(config.Domains))
	for _, d := range config.Domains {
		domainMap[d.Host] = d
	}

	return func(c *fiber.Ctx) error {
		host := cleanHost(c.Get("Host"))
		if host == "localhost" || host == "remu" {
			host = config.LocalhostTestDomain
		}

		domainCfg, exists := domainMap[host]
		if !exists {
			return c.Next()
		}

		if domainCfg.APIEnabled && strings.HasPrefix(c.Path(), "/api/") {
			return c.Next()
		}

		pathKey := host + c.Path()
		if c.Path() == "/" || strings.HasSuffix(c.Path(), "/") {
			pathKey = host + c.Path() + "index.html"
		}

		if cf, ok := cache[pathKey]; ok {
			c.Set(fiber.HeaderContentType, cf.contentType)
			setCacheHeaders(c, c.Path(), config)
			return c.Send(cf.data)
		}

		indexKey := host + "/index.html"
		if cf, ok := cache[indexKey]; ok {
			c.Set(fiber.HeaderContentType, cf.contentType)
			setCacheHeaders(c, "/index.html", config)
			return c.Send(cf.data)
		}

		return c.Next()
	}
}
