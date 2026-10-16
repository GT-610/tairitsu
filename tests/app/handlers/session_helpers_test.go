package handlers

import (
	"net/http"
	"testing"
	"time"

	"github.com/GT-610/tairitsu/internal/app/middleware"
	"github.com/GT-610/tairitsu/internal/app/models"
	appservices "github.com/GT-610/tairitsu/internal/app/services"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

// testRequestTimeout replaces the one second budget fiber applies to Test by
// default. An endpoint that hashes a password exceeds it: under the race
// detector bcrypt costs roughly ten times what it does otherwise, so a login or
// a registration on a loaded machine would fail with an i/o timeout instead of
// reporting what the handler actually did. The Go test deadline stays as the
// backstop for a handler that really does hang.
var testRequestTimeout = fiber.TestConfig{Timeout: 10 * time.Second, FailOnTimeout: true}

// newTestSession creates a session for a user and returns the raw credential
// that a browser would hold. Tests authenticate by putting it in the session
// cookie, mirroring how a real client proves its identity.
func newTestSession(t *testing.T, sessionService *appservices.SessionService, userID string, userAgent string, ipAddress string) (*models.Session, string) {
	t.Helper()

	session, credential, err := sessionService.CreateSession(appservices.SessionCreateInput{
		UserID:     userID,
		UserAgent:  userAgent,
		IPAddress:  ipAddress,
		RememberMe: true,
		ExpiresAt:  time.Now().Add(time.Hour),
	})
	require.NoError(t, err)

	return session, credential
}

// authenticateRequest attaches the session cookie a browser would send.
func authenticateRequest(req *http.Request, credential string) {
	req.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: credential})
}
