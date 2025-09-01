package main

import (
	"log"
	"mime"
	"net/http"
	"os"
	filePathPkg "path/filepath"
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
		filePathPkg.Walk(domain.StaticPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}

			relPath, _ := filePathPkg.Rel(domain.StaticPath, path)
			relPath = "/" + filePathPkg.ToSlash(relPath)

			ext := filePathPkg.Ext(path)
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

const (
	locIsAPI        = "bf_is_api"
	locServedStatic = "bf_served_static"
	locServedHTML   = "bf_served_html"
	locStaticTry    = "bf_static_attempt"
	locSPA          = "bf_served_spa_fallback"
)

func hasExt(p string) bool {
	seg := p
	if i := strings.LastIndexByte(seg, '/'); i >= 0 {
		seg = seg[i+1:]
	}
	j := strings.LastIndexByte(seg, '.')
	return j > 0
}

func normalizePath(p string) string {
	for strings.Contains(p, "//") {
		p = strings.ReplaceAll(p, "//", "/")
	}
	if p == "" || p[0] != '/' {
		p = "/" + p
	}
	return p
}

func staticFilesMiddleware(config *ServerConfig) fiber.Handler {
	cache := preloadStaticFiles(config)

	domainMap := make(map[string]DomainConfig, len(config.Domains))
	for _, d := range config.Domains {
		domainMap[d.Host] = d
	}

	return func(c *fiber.Ctx) error {
		host := cleanHost(c.Hostname())
		if host == "localhost" || host == "remu" {
			host = config.LocalhostTestDomain
		}
		domainCfg, ok := domainMap[host]
		if !ok {
			return c.Next()
		}

		path := normalizePath(c.Path())

		// 1) API passthrough
		if domainCfg.APIEnabled && strings.HasPrefix(path, "/api/") {
			c.Locals(locIsAPI, true)
			return c.Next()
		}

		// 2) Paths WITH extension -> serve or 404 (no SPA)
		if hasExt(path) {
			pathKey := host + path
			if cf, ok := cache[pathKey]; ok {
				c.Locals(locServedStatic, true)
				if strings.HasSuffix(strings.ToLower(path), ".html") {
					c.Locals(locServedHTML, true)
				}
				c.Set(fiber.HeaderContentType, cf.contentType)
				setCacheHeaders(c, pathKey, config)
				if c.Method() == fiber.MethodHead {
					return c.SendStatus(http.StatusOK)
				}
				return c.Send(cf.data)
			}
			c.Locals(locStaticTry, true)
			return c.SendStatus(http.StatusNotFound)
		}

		// 3) Paths WITHOUT extension
		// 3a) "/" or endswith "/" -> try directory index
		if path == "/" || strings.HasSuffix(path, "/") {
			dirKey := host + path + "index.html"
			if cf, ok := cache[dirKey]; ok {
				c.Locals(locServedStatic, true)
				c.Locals(locServedHTML, true)
				c.Set(fiber.HeaderContentType, cf.contentType)
				setCacheHeaders(c, dirKey, config)
				if c.Method() == fiber.MethodHead {
					return c.SendStatus(http.StatusOK)
				}
				return c.Send(cf.data)
			}
		} else {
			// 3b) plain "<path>.html"
			htmlKey := host + path + ".html"
			if cf, ok := cache[htmlKey]; ok {
				c.Locals(locServedStatic, true)
				c.Locals(locServedHTML, true)
				c.Set(fiber.HeaderContentType, cf.contentType)
				setCacheHeaders(c, htmlKey, config)
				if c.Method() == fiber.MethodHead {
					return c.SendStatus(http.StatusOK)
				}
				return c.Send(cf.data)
			}
		}

		// 3c) SPA fallback to site root "/index.html"
		indexKey := host + "/index.html"
		if cf, ok := cache[indexKey]; ok {
			c.Locals(locSPA, true)
			c.Set(fiber.HeaderContentType, cf.contentType)
			setCacheHeaders(c, indexKey, config)
			if c.Method() == fiber.MethodHead {
				return c.SendStatus(http.StatusOK)
			}
			return c.Send(cf.data)
		}

		return c.SendStatus(http.StatusNotFound)
	}
}
