package middleware

import (
	"strings"

	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/gofiber/fiber/v3"
)

// SetupTokenHeader carries the first-run setup token.
const SetupTokenHeader = "X-Setup-Token"

// LocalsSetupAuthorized marks a request that proved it may see setup-only
// details.
const LocalsSetupAuthorized = "setup_authorized"

// SetupTokenRequired rejects setup requests that do not carry the token printed
// at startup, so the first-run window is not open to everyone who can reach the
// port.
//
// It answers 403 rather than 401 because a 401 is indistinguishable from an
// expired session, and the client clears its authentication state on that
// status.
func SetupTokenRequired(token *services.SetupToken) fiber.Handler {
	return func(c fiber.Ctx) error {
		if !token.Matches(strings.TrimSpace(c.Get(SetupTokenHeader))) {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
				Error:     "Forbidden",
				Message:   "A valid setup token is required to configure an uninitialized instance",
				ErrorCode: "setup.token_required",
			})
		}

		c.Locals(LocalsSetupAuthorized, true)
		return c.Next()
	}
}

// SetupTokenOptional records that the caller presented the setup token without
// requiring one, so an otherwise public endpoint can decide whether it may
// include setup-only details.
func SetupTokenOptional(token *services.SetupToken) fiber.Handler {
	return func(c fiber.Ctx) error {
		if token.Matches(strings.TrimSpace(c.Get(SetupTokenHeader))) {
			c.Locals(LocalsSetupAuthorized, true)
		}

		return c.Next()
	}
}

// SetupAuthorized reports whether the current request proved it may see
// setup-only details.
func SetupAuthorized(c fiber.Ctx) bool {
	authorized, _ := c.Locals(LocalsSetupAuthorized).(bool)
	return authorized
}
