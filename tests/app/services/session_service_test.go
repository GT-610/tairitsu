package services

import (
	"testing"
	"time"

	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionService_ResolveAndRevokeOtherSessions(t *testing.T) {
	db := testutil.NewSQLiteDB(t)

	user := &models.User{
		ID:        "user-1",
		Username:  "alice",
		Password:  "hashed-password",
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, db.CreateUser(user))

	sessionService := services.NewSessionService(db)
	currentSession, currentCredential, err := sessionService.CreateSession(services.SessionCreateInput{
		UserID:     user.ID,
		UserAgent:  "browser-a",
		IPAddress:  "127.0.0.1",
		RememberMe: true,
		ExpiresAt:  time.Now().Add(time.Hour),
	})
	require.NoError(t, err)

	_, otherCredential, err := sessionService.CreateSession(services.SessionCreateInput{
		UserID:     user.ID,
		UserAgent:  "browser-b",
		IPAddress:  "127.0.0.2",
		RememberMe: false,
		ExpiresAt:  time.Now().Add(time.Hour),
	})
	require.NoError(t, err)

	resolved, err := sessionService.ResolveSession(currentCredential)
	require.NoError(t, err)
	assert.Equal(t, currentSession.ID, resolved.ID)

	count, err := sessionService.RevokeOtherSessions(user.ID, currentSession.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	_, err = sessionService.ResolveSession(otherCredential)
	assert.ErrorIs(t, err, services.ErrSessionRevoked)

	stillValid, err := sessionService.ResolveSession(currentCredential)
	require.NoError(t, err)
	assert.Equal(t, currentSession.ID, stillValid.ID)
}

func TestSessionService_ResolveRejectsUnknownCredential(t *testing.T) {
	db := testutil.NewSQLiteDB(t)

	sessionService := services.NewSessionService(db)
	_, err := sessionService.ResolveSession("not-a-real-credential")
	assert.ErrorIs(t, err, services.ErrSessionNotFound)
}

func TestSessionService_StoresOnlyCredentialHash(t *testing.T) {
	db := testutil.NewSQLiteDB(t)

	sessionService := services.NewSessionService(db)
	session, credential, err := sessionService.CreateSession(services.SessionCreateInput{
		UserID:    "user-1",
		UserAgent: "browser-a",
		IPAddress: "127.0.0.1",
		ExpiresAt: time.Now().Add(time.Hour),
	})
	require.NoError(t, err)
	require.NotEmpty(t, credential)

	require.NotNil(t, session.TokenHash)
	// The plaintext credential must never be recoverable from the stored row.
	assert.NotEqual(t, credential, *session.TokenHash)

	stored, err := db.GetSessionByID(session.ID)
	require.NoError(t, err)
	require.NotNil(t, stored.TokenHash)
	assert.NotEqual(t, credential, *stored.TokenHash)
}

func TestSessionService_PurgeUnusableSessionsRemovesCredentiallessRows(t *testing.T) {
	db := testutil.NewSQLiteDB(t)

	now := time.Now()
	// A row written before session cookies existed carries no credential.
	require.NoError(t, db.CreateSession(&models.Session{
		ID:         "legacy-session",
		UserID:     "user-1",
		UserAgent:  "legacy-browser",
		IPAddress:  "127.0.0.1",
		LastSeenAt: now,
		ExpiresAt:  now.Add(time.Hour),
		CreatedAt:  now,
		UpdatedAt:  now,
	}))

	sessionService := services.NewSessionService(db)
	activeSession, credential, err := sessionService.CreateSession(services.SessionCreateInput{
		UserID:    "user-1",
		UserAgent: "browser-a",
		IPAddress: "127.0.0.1",
		ExpiresAt: now.Add(time.Hour),
	})
	require.NoError(t, err)

	sessionService.PurgeUnusableSessions()

	remaining, err := sessionService.GetUserSessions("user-1")
	require.NoError(t, err)
	require.Len(t, remaining, 1)
	assert.Equal(t, activeSession.ID, remaining[0].ID)

	// The surviving session must still authenticate.
	_, err = sessionService.ResolveSession(credential)
	require.NoError(t, err)
}

// The session list is keyed by session ID and the ID travels in the URL, so
// without an ownership check any signed-in account could sign out any other by
// guessing or observing an ID.
func TestSessionService_RevokeRejectsAnotherUsersSession(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	now := time.Now()
	require.NoError(t, db.CreateUser(&models.User{
		ID: "user-1", Username: "alice", Password: "hashed", Role: "user", CreatedAt: now, UpdatedAt: now,
	}))
	require.NoError(t, db.CreateUser(&models.User{
		ID: "user-2", Username: "bob", Password: "hashed", Role: "user", CreatedAt: now, UpdatedAt: now,
	}))

	sessionService := services.NewSessionService(db)

	victimSession, victimCredential, err := sessionService.CreateSession(services.SessionCreateInput{
		UserID:     "user-1",
		UserAgent:  "victim-browser",
		IPAddress:  "127.0.0.1",
		RememberMe: true,
		ExpiresAt:  now.Add(time.Hour),
	})
	require.NoError(t, err)

	// Bob is signed in on his own account and names Alice's session ID.
	err = sessionService.RevokeSession("user-2", victimSession.ID)
	assert.ErrorIs(t, err, services.ErrSessionAccessDenied)

	// The credential must still work: a refused revocation cannot have revoked.
	_, err = sessionService.ResolveSession(victimCredential)
	assert.NoError(t, err, "another account must not be able to end this session")

	// The owner can, so the refusal is about ownership and not a broken call.
	require.NoError(t, sessionService.RevokeSession("user-1", victimSession.ID))
	_, err = sessionService.ResolveSession(victimCredential)
	assert.ErrorIs(t, err, services.ErrSessionRevoked)
}

// An unknown session ID is a missing resource rather than a permission problem,
// which is what the route turns into a 404.
func TestSessionService_RevokeRejectsUnknownSession(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	sessionService := services.NewSessionService(db)

	err := sessionService.RevokeSession("user-1", "no-such-session")
	assert.ErrorIs(t, err, services.ErrSessionNotFound)
}
