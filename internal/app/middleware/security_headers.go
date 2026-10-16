package middleware

import "github.com/gofiber/fiber/v3"

func SecurityHeaders() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; base-uri 'self'; frame-ancestors 'none'; object-src 'none'")
		c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		if c.Scheme() == "https" {
			c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		return c.Next()
	}
}

// NoStore marks responses as uncacheable.
//
// It belongs on the JSON surface rather than on the whole process: the
// console's hashed assets are meant to be cached for a year, while an API
// response can carry a session's data and two of them carry a one-time
// password. Nothing downstream should keep a copy of those.
func NoStore() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Set(fiber.HeaderCacheControl, "no-store")
		return c.Next()
	}
}
