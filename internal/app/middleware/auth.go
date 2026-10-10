package middleware

import (
	"errors"
	"time"

	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// SessionCookieName is the cookie that carries the opaque session credential.
// It is HttpOnly, so frontend code cannot read it and a script injection cannot
// exfiltrate it.
const SessionCookieName = "tairitsu_session"

// sessionCookiePath scopes the cookie to the whole application. The frontend is
// served from the same origin as the API, so a narrower path is unnecessary.
const sessionCookiePath = "/"

// SessionCookie builds the cookie for a freshly issued session credential.
// persistent controls whether the cookie survives a browser restart, matching
// the login form's "remember me" choice. The server-side session expiry is
// authoritative either way.
func SessionCookie(credential string, expiresAt time.Time, persistent bool, secure bool) *fiber.Cookie {
	cookie := &fiber.Cookie{
		Name:     SessionCookieName,
		Value:    credential,
		Path:     sessionCookiePath,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: fiber.CookieSameSiteLaxMode,
	}

	if persistent {
		cookie.Expires = expiresAt
		if remaining := int(time.Until(expiresAt).Seconds()); remaining > 0 {
			cookie.MaxAge = remaining
		}
	} else {
		cookie.SessionOnly = true
	}

	return cookie
}

// ExpiredSessionCookie clears the session cookie on the client.
func ExpiredSessionCookie(secure bool) *fiber.Cookie {
	return &fiber.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     sessionCookiePath,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: fiber.CookieSameSiteLaxMode,
		Expires:  time.Unix(1, 0),
		MaxAge:   -1,
	}
}

// IsSecureRequest reports whether the browser reached the server over HTTPS.
// TLS is commonly terminated by a reverse proxy, so the forwarded scheme counts.
func IsSecureRequest(c fiber.Ctx) bool {
	return c.Secure() || c.Protocol() == "https"
}

// AuthMiddleware authenticates a request from its session cookie and loads the
// owning user, so handlers work from a verified identity instead of trusting
// client-supplied claims.
func AuthMiddleware(sessionService *services.SessionService, userService *services.UserService) fiber.Handler {
	return func(c fiber.Ctx) error {
		credential := c.Cookies(SessionCookieName)
		if credential == "" {
			return unauthorized(c, "auth.missing_token", "Authentication required")
		}

		if sessionService == nil || userService == nil {
			logger.Error("authentication is unavailable because required services are missing")
			return c.Status(fiber.StatusServiceUnavailable).JSON(ErrorResponse{
				Error:     "Service Unavailable",
				Message:   "Authentication service is unavailable",
				ErrorCode: "system.auth_unavailable",
				Code:      fiber.StatusServiceUnavailable,
			})
		}

		session, err := sessionService.ResolveSession(credential)
		if err != nil {
			if errors.Is(err, services.ErrUserDBUnavailable) {
				logger.Error("authentication failed because the user database is unavailable", zap.Error(err))
				return c.Status(fiber.StatusServiceUnavailable).JSON(ErrorResponse{
					Error:     "Service Unavailable",
					Message:   "Authentication service is unavailable",
					ErrorCode: "user.db_unavailable",
					Code:      fiber.StatusServiceUnavailable,
				})
			}
			return unauthorized(c, "auth.invalid_token", "Invalid or expired session")
		}

		user, err := userService.GetUserByID(session.UserID)
		if err != nil {
			if services.IsUserDBUnavailable(err) {
				logger.Error("authentication failed because the user database is unavailable", zap.Error(err))
				return c.Status(fiber.StatusServiceUnavailable).JSON(ErrorResponse{
					Error:     "Service Unavailable",
					Message:   "Authentication service is unavailable",
					ErrorCode: "user.db_unavailable",
					Code:      fiber.StatusServiceUnavailable,
				})
			}
			return unauthorized(c, "auth.invalid_token", "Invalid or expired session")
		}
		if user == nil {
			// The session outlived its user. Clearing the cookie here stops the
			// browser from replaying a credential that can never succeed.
			return unauthorized(c, "auth.invalid_token", "Invalid or expired session")
		}

		if err := sessionService.TouchSession(session); err != nil {
			logger.Warn("failed to record session activity", zap.String("session_id", session.ID), zap.Error(err))
		}

		c.Locals("user_id", user.ID)
		c.Locals("username", user.Username)
		c.Locals("role", user.Role)
		c.Locals("session_id", session.ID)

		return c.Next()
	}
}

func unauthorized(c fiber.Ctx, code string, message string) error {
	c.Cookie(ExpiredSessionCookie(IsSecureRequest(c)))
	return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
		Error:     "Unauthorized",
		Message:   message,
		ErrorCode: code,
		Code:      fiber.StatusUnauthorized,
	})
}

// AdminRequiredWithUserService is the admin authorization middleware.
// It checks the database on every request to detect stale tokens after admin transfers.
func AdminRequiredWithUserService(userService *services.UserService) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, exists := c.Locals("user_id").(string)
		if !exists || userID == "" {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
				Error:     "Forbidden",
				Message:   "Authentication required",
				ErrorCode: "auth.required",
				Code:      fiber.StatusForbidden,
			})
		}

		if userService == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(ErrorResponse{
				Error:     "Service Unavailable",
				Message:   "User service is unavailable",
				ErrorCode: "system.user_service_unavailable",
				Code:      fiber.StatusServiceUnavailable,
			})
		}

		user, err := userService.GetUserByID(userID)
		if err != nil {
			if services.IsUserDBUnavailable(err) {
				logger.Error("Administrator authorization failed because user database is unavailable", zap.Error(err))
				return c.Status(fiber.StatusServiceUnavailable).JSON(ErrorResponse{
					Error:     "Service Unavailable",
					Message:   "User service is unavailable",
					ErrorCode: "user.db_unavailable",
					Code:      fiber.StatusServiceUnavailable,
				})
			}
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
				Error:     "Forbidden",
				Message:   "Administrator permission required",
				ErrorCode: "auth.admin_required",
				Code:      fiber.StatusForbidden,
			})
		}

		if user.Role != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
				Error:     "Forbidden",
				Message:   "Administrator permission required",
				ErrorCode: "auth.admin_required",
				Code:      fiber.StatusForbidden,
			})
		}

		c.Locals("role", user.Role)
		return c.Next()
	}
}
