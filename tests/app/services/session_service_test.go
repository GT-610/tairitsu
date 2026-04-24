package services

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/GT-610/tairitsu/internal/app/database"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionService_CreateValidateAndRevokeOtherSessions(t *testing.T) {
	db, err := database.NewSQLiteDB(filepath.Join(t.TempDir(), "tairitsu.db"))
	require.NoError(t, err)
	require.NoError(t, db.Init())
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

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
	currentSession, err := sessionService.CreateSession(services.SessionCreateInput{
		UserID:     user.ID,
		UserAgent:  "browser-a",
		IPAddress:  "127.0.0.1",
		RememberMe: true,
		ExpiresAt:  time.Now().Add(time.Hour),
	})
	require.NoError(t, err)

	otherSession, err := sessionService.CreateSession(services.SessionCreateInput{
		UserID:     user.ID,
		UserAgent:  "browser-b",
		IPAddress:  "127.0.0.2",
		RememberMe: false,
		ExpiresAt:  time.Now().Add(time.Hour),
	})
	require.NoError(t, err)

	validated, err := sessionService.ValidateSession(user.ID, currentSession.ID)
	require.NoError(t, err)
	assert.Equal(t, currentSession.ID, validated.ID)

	count, err := sessionService.RevokeOtherSessions(user.ID, currentSession.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	_, err = sessionService.ValidateSession(user.ID, otherSession.ID)
	assert.ErrorIs(t, err, services.ErrSessionRevoked)

	stillValid, err := sessionService.ValidateSession(user.ID, currentSession.ID)
	require.NoError(t, err)
	assert.Equal(t, currentSession.ID, stillValid.ID)
}
