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

// TestUserService_LoginThrottlesAfterRepeatedFailures is the regression test for
// unlimited password guessing: the transport limiter keys on an address the
// caller can choose when a proxy forwards their own X-Forwarded-For, so the
// account is what has to be counted.
func TestUserService_LoginThrottlesAfterRepeatedFailures(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	userService := services.NewUserService(db)

	_, err := userService.Register(&models.RegisterRequest{Username: "alice", Password: "secret123"}, "user")
	require.NoError(t, err)

	// Burn through the allowance.
	for attempt := 0; attempt < 6; attempt++ {
		_, loginErr := userService.Login(&models.LoginRequest{Username: "alice", Password: "wrong"})
		require.ErrorIs(t, loginErr, services.ErrInvalidCredentials)
	}

	// The next attempt is refused without the password being examined at all.
	_, err = userService.Login(&models.LoginRequest{Username: "alice", Password: "wrong"})
	require.ErrorIs(t, err, services.ErrLoginThrottled)

	retryAfter := services.LoginRetryAfter(err)
	assert.Positive(t, retryAfter, "a throttled sign-in must say how long to wait")
	assert.LessOrEqual(t, retryAfter, 5*time.Minute)

	// A correct password is refused too: the point is to stop guessing, not to
	// let a lucky guess through.
	_, err = userService.Login(&models.LoginRequest{Username: "alice", Password: "secret123"})
	require.ErrorIs(t, err, services.ErrLoginThrottled)
}

// TestUserService_LoginThrottleForgetsOnSuccess keeps the throttle from
// punishing a user who eventually remembers their password.
func TestUserService_LoginThrottleForgetsOnSuccess(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	userService := services.NewUserService(db)

	_, err := userService.Register(&models.RegisterRequest{Username: "alice", Password: "secret123"}, "user")
	require.NoError(t, err)

	for attempt := 0; attempt < 4; attempt++ {
		_, loginErr := userService.Login(&models.LoginRequest{Username: "alice", Password: "wrong"})
		require.ErrorIs(t, loginErr, services.ErrInvalidCredentials)
	}

	_, err = userService.Login(&models.LoginRequest{Username: "alice", Password: "secret123"})
	require.NoError(t, err)

	// The count restarted, so four more misses are still free.
	for attempt := 0; attempt < 4; attempt++ {
		_, loginErr := userService.Login(&models.LoginRequest{Username: "alice", Password: "wrong"})
		require.ErrorIs(t, loginErr, services.ErrInvalidCredentials)
	}
}

// TestUserService_LoginThrottleDoesNotRevealWhetherAnAccountExists is the
// username-enumeration guard: an unknown account must be throttled exactly like a
// known one, and must produce the same error before and after throttling.
func TestUserService_LoginThrottleDoesNotRevealWhetherAnAccountExists(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	userService := services.NewUserService(db)

	_, err := userService.Register(&models.RegisterRequest{Username: "alice", Password: "secret123"}, "user")
	require.NoError(t, err)

	// Both names fail identically before throttling.
	_, knownErr := userService.Login(&models.LoginRequest{Username: "alice", Password: "wrong"})
	_, unknownErr := userService.Login(&models.LoginRequest{Username: "nobody", Password: "wrong"})
	require.ErrorIs(t, knownErr, services.ErrInvalidCredentials)
	require.ErrorIs(t, unknownErr, services.ErrInvalidCredentials)
	assert.Equal(t, knownErr.Error(), unknownErr.Error())

	// Drive both past the allowance and confirm both land on the same refusal.
	for attempt := 0; attempt < 6; attempt++ {
		_, _ = userService.Login(&models.LoginRequest{Username: "alice", Password: "wrong"})
		_, _ = userService.Login(&models.LoginRequest{Username: "nobody", Password: "wrong"})
	}

	_, knownErr = userService.Login(&models.LoginRequest{Username: "alice", Password: "wrong"})
	_, unknownErr = userService.Login(&models.LoginRequest{Username: "nobody", Password: "wrong"})

	require.ErrorIs(t, knownErr, services.ErrLoginThrottled)
	require.ErrorIs(t, unknownErr, services.ErrLoginThrottled)
	assert.Equal(t, knownErr.Error(), unknownErr.Error())
}

// TestUserService_LoginThrottleFoldsCase stops a caller from buying a fresh
// allowance by respelling the same account.
func TestUserService_LoginThrottleFoldsCase(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	userService := services.NewUserService(db)

	_, err := userService.Register(&models.RegisterRequest{Username: "alice", Password: "secret123"}, "user")
	require.NoError(t, err)

	for attempt := 0; attempt < 6; attempt++ {
		_, _ = userService.Login(&models.LoginRequest{Username: "alice", Password: "wrong"})
	}

	_, err = userService.Login(&models.LoginRequest{Username: "ALICE", Password: "wrong"})
	require.ErrorIs(t, err, services.ErrLoginThrottled)
}
