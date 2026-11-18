package services

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/GT-610/tempest-zt/internal/app/models"
	"github.com/GT-610/tempest-zt/internal/app/services"
	"github.com/GT-610/tempest-zt/tests/testutil"
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

// TestUserService_LoginThrottleSeparatesAccountsThatDifferOnlyByCase is the
// regression test for sharing one record between two accounts. Account lookup is
// case-sensitive, so "alice" and "ALICE" can both exist; a shared record let
// whoever holds the second clear the first's backoff by signing in.
func TestUserService_LoginThrottleSeparatesAccountsThatDifferOnlyByCase(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	userService := services.NewUserService(db)

	_, err := userService.Register(&models.RegisterRequest{Username: "alice", Password: "secret123"}, "user")
	require.NoError(t, err)
	_, err = userService.Register(&models.RegisterRequest{Username: "ALICE", Password: "other12345"}, "user")
	require.NoError(t, err)

	// Block alice.
	for attempt := 0; attempt < 6; attempt++ {
		_, _ = userService.Login(&models.LoginRequest{Username: "alice", Password: "wrong"})
	}
	_, err = userService.Login(&models.LoginRequest{Username: "alice", Password: "wrong"})
	require.ErrorIs(t, err, services.ErrLoginThrottled)

	// Signing in to the other account must not clear alice's backoff.
	_, err = userService.Login(&models.LoginRequest{Username: "ALICE", Password: "other12345"})
	require.NoError(t, err, "the other account is not blocked")

	_, err = userService.Login(&models.LoginRequest{Username: "alice", Password: "wrong"})
	require.ErrorIs(t, err, services.ErrLoginThrottled, "the other account's success must not clear alice's block")
}

// TestUserService_LoginThrottleOneVerificationAtATime covers the parallel case: a
// burst for one account must not each reach the password comparison before any of
// them has recorded a failure.
func TestUserService_LoginThrottleOneVerificationAtATime(t *testing.T) {
	db := testutil.NewSQLiteDB(t)
	userService := services.NewUserService(db)

	_, err := userService.Register(&models.RegisterRequest{Username: "alice", Password: "secret123"}, "user")
	require.NoError(t, err)

	// Occupy the account the way an in-progress verification would.
	//
	// This goes through the exported surface only, so it drives real logins: some
	// of the burst is refused rather than all of it being served.
	var mutex sync.Mutex
	throttled := 0
	var waitGroup sync.WaitGroup
	for attempt := 0; attempt < 30; attempt++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			_, loginErr := userService.Login(&models.LoginRequest{Username: "alice", Password: "wrong"})
			if errors.Is(loginErr, services.ErrLoginThrottled) {
				mutex.Lock()
				throttled++
				mutex.Unlock()
			}
		}()
	}
	waitGroup.Wait()

	assert.Positive(t, throttled, "a burst must be partly refused rather than all served")
}
