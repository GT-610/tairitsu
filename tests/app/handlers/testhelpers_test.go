package handlers

import (
	"net/http"
	"testing"
	"time"

	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/middleware"
	"github.com/GT-610/tairitsu/internal/app/models"
	appservices "github.com/GT-610/tairitsu/internal/app/services"
	"github.com/stretchr/testify/require"
)

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

// newTestDatabase opens an isolated SQLite database for a test.
func newTestDatabase(t *testing.T) database.DBInterface {
	t.Helper()

	db, err := database.NewDatabase(database.Config{
		Type: database.SQLite,
		Path: t.TempDir() + "/tairitsu.db",
	})
	require.NoError(t, err)
	require.NoError(t, db.Init())
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	return db
}
