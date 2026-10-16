package middleware

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/middleware"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/tests/testutil"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sessionCookieName mirrors the cookie the login handler issues.
const sessionCookieName = middleware.SessionCookieName

func createTestUser(t *testing.T, db database.DBInterface, user *models.User) {
	t.Helper()

	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = time.Now()
	}
	if user.Password == "" {
		user.Password = "hashed-password"
	}
	require.NoError(t, db.CreateUser(user))
}

// issueSession creates a session and returns the browser credential.
func issueSession(t *testing.T, sessionService *services.SessionService, userID string) string {
	t.Helper()

	_, credential, err := sessionService.CreateSession(services.SessionCreateInput{
		UserID:     userID,
		UserAgent:  "test-agent",
		IPAddress:  "127.0.0.1",
		RememberMe: true,
		ExpiresAt:  time.Now().Add(time.Hour),
	})
	require.NoError(t, err)

	return credential
}

func TestAuthMiddleware_MissingCookie(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	sessionService := services.NewSessionService(db)
	userService := services.NewUserService(db)

	router := fiber.New()
	router.Use(middleware.AuthMiddleware(sessionService, userService))
	router.Get("/test", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := router.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Contains(t, string(body), "auth.missing_token")
}

func TestAuthMiddleware_RejectsUnknownCredential(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	sessionService := services.NewSessionService(db)
	userService := services.NewUserService(db)

	router := fiber.New()
	router.Use(middleware.AuthMiddleware(sessionService, userService))
	router.Get("/test", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "forged-credential"})
	resp, err := router.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Contains(t, string(body), "auth.invalid_token")
}

func TestAuthMiddleware_AcceptsValidSessionCookie(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	sessionService := services.NewSessionService(db)
	userService := services.NewUserService(db)

	user := &models.User{ID: "test-user-id", Username: "test-user", Role: "user"}
	createTestUser(t, db, user)
	credential := issueSession(t, sessionService, user.ID)

	router := fiber.New()
	router.Use(middleware.AuthMiddleware(sessionService, userService))
	router.Get("/test", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"user_id":  c.Locals("user_id"),
			"username": c.Locals("username"),
			"role":     c.Locals("role"),
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: credential})
	resp, err := router.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Contains(t, string(body), "test-user-id")
	assert.Contains(t, string(body), "test-user")
}

func TestAuthMiddleware_RejectsCredentialAfterLogout(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	sessionService := services.NewSessionService(db)
	userService := services.NewUserService(db)

	user := &models.User{ID: "user-1", Username: "alice", Role: "user"}
	createTestUser(t, db, user)

	session, credential, err := sessionService.CreateSession(services.SessionCreateInput{
		UserID:    user.ID,
		UserAgent: "test-agent",
		IPAddress: "127.0.0.1",
		ExpiresAt: time.Now().Add(time.Hour),
	})
	require.NoError(t, err)
	require.NoError(t, sessionService.RevokeSession(user.ID, session.ID))

	router := fiber.New()
	router.Use(middleware.AuthMiddleware(sessionService, userService))
	router.Get("/test", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: credential})
	resp, err := router.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAuthMiddleware_ClearsCookieWhenUserIsGone(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	sessionService := services.NewSessionService(db)
	userService := services.NewUserService(db)

	// A session whose user no longer exists must not keep working, and the
	// browser should be told to drop the credential.
	credential := issueSession(t, sessionService, "deleted-user")

	router := fiber.New()
	router.Use(middleware.AuthMiddleware(sessionService, userService))
	router.Get("/test", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: credential})
	resp, err := router.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	cleared := false
	for _, cookie := range resp.Cookies() {
		if cookie.Name == sessionCookieName && cookie.Value == "" {
			cleared = true
		}
	}
	assert.True(t, cleared, "expected the session cookie to be cleared")
}

// When the session lookup fails for a reason that does not prove the credential
// is bad, authentication must fail closed without discarding it. Clearing the
// cookie here would sign the user out over a passing fault.
func TestAuthMiddleware_KeepsCookieWhenSessionLookupFails(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	sessionService := services.NewSessionService(db)
	userService := services.NewUserService(db)

	user := &models.User{ID: "user-1", Username: "alice", Role: "user"}
	createTestUser(t, db, user)
	credential := issueSession(t, sessionService, user.ID)

	// Closing the database makes the lookup fail with a read error that is not
	// one of the "credential is invalid" sentinels.
	require.NoError(t, db.Close())

	router := fiber.New()
	router.Use(middleware.AuthMiddleware(sessionService, userService))
	router.Get("/test", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: credential})
	resp, err := router.Test(req)

	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
	for _, cookie := range resp.Cookies() {
		if cookie.Name == sessionCookieName {
			assert.NotEmpty(t, cookie.Value, "a transient failure must not clear the session cookie")
		}
	}

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "user.db_unavailable")
}

// A session that is genuinely expired must still clear the cookie, so the two
// outcomes stay distinct.
func TestAuthMiddleware_ClearsCookieWhenSessionExpired(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	sessionService := services.NewSessionService(db)
	userService := services.NewUserService(db)

	user := &models.User{ID: "user-1", Username: "alice", Role: "user"}
	createTestUser(t, db, user)

	_, credential, err := sessionService.CreateSession(services.SessionCreateInput{
		UserID:    user.ID,
		UserAgent: "test-agent",
		IPAddress: "127.0.0.1",
		ExpiresAt: time.Now().Add(-time.Minute),
	})
	require.NoError(t, err)

	router := fiber.New()
	router.Use(middleware.AuthMiddleware(sessionService, userService))
	router.Get("/test", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: credential})
	resp, err := router.Test(req)

	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	cleared := false
	for _, cookie := range resp.Cookies() {
		if cookie.Name == sessionCookieName && cookie.Value == "" {
			cleared = true
		}
	}
	assert.True(t, cleared, "an expired session must clear the cookie")
}

func TestAdminRequiredWithUserService_DeniesStaleAdminSessionAfterTransfer(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	userService := services.NewUserService(db)
	sessionService := services.NewSessionService(db)

	currentAdmin := &models.User{ID: "admin-1", Username: "admin", Role: "user"}
	createTestUser(t, db, currentAdmin)
	credential := issueSession(t, sessionService, currentAdmin.ID)

	router := fiber.New()
	router.Use(middleware.AuthMiddleware(sessionService, userService))
	router.Use(middleware.AdminRequiredWithUserService(userService))
	router.Get("/admin", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "admin access granted"})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: credential})
	resp, err := router.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestSessionCookieIsHttpOnlyAndSameSiteLax(t *testing.T) {
	expiresAt := time.Now().Add(time.Hour)

	persistent := middleware.SessionCookie("credential", expiresAt, true, false)
	assert.True(t, persistent.HTTPOnly)
	assert.Equal(t, fiber.CookieSameSiteLaxMode, persistent.SameSite)
	assert.Equal(t, "/", persistent.Path)
	// Plain HTTP must not set Secure, or the cookie would never be sent.
	assert.False(t, persistent.Secure)
	assert.Positive(t, persistent.MaxAge)

	transient := middleware.SessionCookie("credential", expiresAt, false, false)
	assert.True(t, transient.SessionOnly)
	assert.Zero(t, transient.MaxAge)

	secure := middleware.SessionCookie("credential", expiresAt, true, true)
	assert.True(t, secure.Secure)
}

func TestExpiredSessionCookieClearsTheCredential(t *testing.T) {
	cookie := middleware.ExpiredSessionCookie(false)

	assert.Equal(t, middleware.SessionCookieName, cookie.Name)
	assert.Empty(t, cookie.Value)
	assert.Negative(t, cookie.MaxAge)
	assert.True(t, cookie.HTTPOnly)
}

// A process that could not assemble its services must refuse the request rather
// than let it through unauthenticated. Passing nil is how assembly reports that
// the pieces are missing.
func TestAuthMiddleware_RefusesWhenServicesAreMissing(t *testing.T) {
	user := &models.User{ID: "user-1", Username: "alice", Role: "user"}

	cases := []struct {
		name           string
		sessionService *services.SessionService
		userService    *services.UserService
	}{
		{name: "no session service", sessionService: nil, userService: services.NewUserService(testutil.NewFakeDB(user))},
		{name: "no user service", sessionService: services.NewSessionService(testutil.NewFakeDB(user)), userService: nil},
		{name: "neither service", sessionService: nil, userService: nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			router := fiber.New()
			router.Use(middleware.AuthMiddleware(testCase.sessionService, testCase.userService))
			router.Get("/test", func(c fiber.Ctx) error {
				return c.JSON(fiber.Map{"message": "success"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "some-credential"})
			resp, err := router.Test(req)

			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Contains(t, string(body), "system.auth_unavailable")
		})
	}
}

// Reading the session user can fail without the credential being bad, for
// instance when the database is unreachable. That has to fail closed and keep
// the cookie, exactly like a failed session lookup.
func TestAuthMiddleware_KeepsCookieWhenUserLookupFails(t *testing.T) {
	user := &models.User{ID: "user-1", Username: "alice", Role: "user"}
	failingDB := testutil.NewFakeDB(user)
	failingDB.GetUserByIDError = errors.New("database is closed")

	sessionService := services.NewSessionService(failingDB)
	userService := services.NewUserService(failingDB)
	credential := issueSession(t, sessionService, user.ID)

	router := fiber.New()
	router.Use(middleware.AuthMiddleware(sessionService, userService))
	router.Get("/test", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: credential})
	resp, err := router.Test(req)

	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
	for _, cookie := range resp.Cookies() {
		if cookie.Name == sessionCookieName {
			assert.NotEmpty(t, cookie.Value, "a transient failure must not clear the session cookie")
		}
	}

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "user.db_unavailable")
}

// IsSecureRequest decides the Secure flag on the session cookie, so a browser
// that reached the server over plain HTTP would silently stop sending it if the
// answer were wrong.
//
// The requests go over a real loopback listener because the answer depends on
// the peer address: fiber's test transport reports 0.0.0.0, which no trusted
// proxy configuration matches, so it can never reach the forwarded-scheme path.
func TestIsSecureRequestFollowsTheRequestScheme(t *testing.T) {
	cases := []struct {
		name           string
		forwardedProto string
		wantSecure     bool
	}{
		{name: "plain http", forwardedProto: "", wantSecure: false},
		{name: "https forwarded by a trusted proxy", forwardedProto: "https", wantSecure: true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			require.NoError(t, err)

			app := fiber.New(fiber.Config{
				TrustProxy:       true,
				TrustProxyConfig: fiber.TrustProxyConfig{Loopback: true},
			})
			app.Get("/test", func(c fiber.Ctx) error {
				return c.JSON(fiber.Map{"secure": middleware.IsSecureRequest(c)})
			})

			serveErr := make(chan error, 1)
			go func() {
				serveErr <- app.Listener(listener)
			}()
			t.Cleanup(func() {
				require.NoError(t, app.Shutdown())
				<-serveErr
			})

			req, err := http.NewRequest(http.MethodGet, "http://"+listener.Addr().String()+"/test", nil)
			require.NoError(t, err)
			if testCase.forwardedProto != "" {
				req.Header.Set(fiber.HeaderXForwardedProto, testCase.forwardedProto)
			}

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			var body struct {
				Secure bool `json:"secure"`
			}
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
			assert.Equal(t, testCase.wantSecure, body.Secure)
		})
	}
}

// A forwarded scheme is a client-supplied header, so it may only be believed
// when the request came from a proxy the operator declared. A deployment that
// trusted it unconditionally would let any caller convince the server it was
// reached over TLS and receive a Secure cookie that the browser then drops.
func TestIsSecureRequestIgnoresAForwardedSchemeFromAnUntrustedPeer(t *testing.T) {
	router := fiber.New()
	router.Get("/test", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"secure": middleware.IsSecureRequest(c)})
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.test/test", nil)
	req.Header.Set(fiber.HeaderXForwardedProto, "https")
	resp, err := router.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var body struct {
		Secure bool `json:"secure"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.False(t, body.Secure)
}
