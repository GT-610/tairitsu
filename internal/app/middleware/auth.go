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
// TLS is commonly terminated by a reverse proxy, so the forwarded scheme counts
// as soon as the proxy itself is trusted.
//
// Fiber exposes the scheme through Secure()/Scheme(). Its Protocol() returns
// the HTTP version ("HTTP/1.1"), so a comparison against Protocol() can never
// be true and must not be used here.
func IsSecureRequest(c fiber.Ctx) bool {
	return c.Secure()
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
			return authUnavailable(c, "system.auth_unavailable", "Authentication service is unavailable")
		}

		session, err := sessionService.ResolveSession(credential)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrSessionNotFound),
				errors.Is(err, services.ErrSessionRevoked),
				errors.Is(err, services.ErrSessionExpired):
				// The credential itself is spent, so dropping it is correct.
				return unauthorized(c, "auth.invalid_token", "Invalid or expired session")
			default:
				// The lookup failed for another reason (database unreachable,
				// transient read error). The credential may still be valid, so
				// fail closed without discarding it; clearing the cookie here
				// would sign the user out over a passing fault.
				logger.Error("authentication failed while resolving the session", zap.Error(err))
				return authUnavailable(c, "user.db_unavailable", "Authentication is temporarily unavailable")
			}
		}

		user, err := userService.GetUserByID(session.UserID)
		if err != nil {
			if errors.Is(err, services.ErrUserNotFound) {
				// The session outlived its user. Clearing the cookie stops the
				// browser replaying a credential that can never succeed.
				return unauthorized(c, "auth.invalid_token", "Invalid or expired session")
			}
			logger.Error("authentication failed while loading the session user", zap.Error(err))
			return authUnavailable(c, "user.db_unavailable", "Authentication is temporarily unavailable")
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
	})
}

// authUnavailable reports that authentication could not be completed. It
// deliberately leaves the session cookie in place: the credential was not
// proven invalid, so the client should retry once the fault passes.
func authUnavailable(c fiber.Ctx, code string, message string) error {
	return c.Status(fiber.StatusServiceUnavailable).JSON(ErrorResponse{
		Error:     "Service Unavailable",
		Message:   message,
		ErrorCode: code,
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
			})
		}

		if userService == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(ErrorResponse{
				Error:     "Service Unavailable",
				Message:   "User service is unavailable",
				ErrorCode: "system.user_service_unavailable",
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
				})
			}
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
				Error:     "Forbidden",
				Message:   "Administrator permission required",
				ErrorCode: "auth.admin_required",
			})
		}

		if user.Role != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
				Error:     "Forbidden",
				Message:   "Administrator permission required",
				ErrorCode: "auth.admin_required",
			})
		}

		c.Locals("role", user.Role)
		return c.Next()
	}
}
