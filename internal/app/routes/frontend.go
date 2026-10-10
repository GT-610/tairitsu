package routes

import (
	"io/fs"
	"strings"

	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/web"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
	"go.uber.org/zap"
)

const (
	// apiPrefix marks paths that must never fall back to the app shell.
	apiPrefix = "/api"
	// assetMaxAge caches hashed build assets for a year. index.html is exempt
	// so a new release is picked up on the next page load.
	assetMaxAge = 31536000
)

// isAPIPath reports whether a request path belongs to the JSON surface.
func isAPIPath(path string) bool {
	return path == apiPrefix || strings.HasPrefix(path, apiPrefix+"/")
}

// mountFrontend serves the embedded frontend bundle from the same process as
// the API. Unknown non-API paths return index.html so client-side routes survive
// a page reload or a direct link.
func mountFrontend(router *fiber.App) {
	assets, err := web.Assets()
	if err != nil {
		logger.Fatal("failed to open embedded frontend assets", zap.Error(err))
	}

	indexHTML, err := fs.ReadFile(assets, "index.html")
	if err != nil {
		// A placeholder bundle is embedded when the frontend was not built
		// before the server. Serving the API alone keeps development and tests
		// usable; point the developer at the build step instead of crashing.
		logger.Warn("embedded frontend bundle is missing index.html; build the frontend to serve the web console")
		router.Use(func(c fiber.Ctx) error {
			if isAPIPath(c.Path()) {
				return c.Next()
			}
			c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
			return c.Status(fiber.StatusServiceUnavailable).SendString(frontendNotBuiltPage)
		})
		return
	}

	serveIndex := func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
		c.Set(fiber.HeaderCacheControl, "no-cache")
		return c.Status(fiber.StatusOK).Send(indexHTML)
	}

	router.Use(static.New("", static.Config{
		FS:     assets,
		MaxAge: assetMaxAge,
		// A successful file response must not carry the asset cache policy when
		// the resolved file is the app shell.
		ModifyResponse: func(c fiber.Ctx) error {
			if strings.HasSuffix(c.Path(), "/") || c.Path() == "/index.html" {
				c.Set(fiber.HeaderCacheControl, "no-cache")
			}
			return nil
		},
		NotFoundHandler: func(c fiber.Ctx) error {
			if isAPIPath(c.Path()) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"message":    "Not Found",
					"error_code": "system.not_found",
					"code":       fiber.StatusNotFound,
				})
			}
			return serveIndex(c)
		},
	}))
}

const frontendNotBuiltPage = `<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><title>Tairitsu</title></head>
<body style="margin: 40px; font-family: sans-serif; color: #333">
<h1>Tairitsu</h1>
<p>The frontend bundle was not built into this binary.</p>
<p>Run <code>bun run build</code> in <code>web/</code>, then rebuild the server.</p>
</body>
</html>`
