package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoginBackoffGrowsThenCaps pins the shape of the delay: nothing up to the
// allowance, then doubling, then a ceiling so an account recovers on its own.
func TestLoginBackoffGrowsThenCaps(t *testing.T) {
	// Up to the allowance a caller pays nothing: mistyping a password a few times
	// is ordinary.
	for failures := 0; failures <= loginFailureAllowance; failures++ {
		assert.Zero(t, loginBackoffFor(failures), "failures=%d", failures)
	}

	// Past it, each failure doubles the wait.
	assert.Equal(t, 2*time.Second, loginBackoffFor(loginFailureAllowance+1))
	assert.Equal(t, 4*time.Second, loginBackoffFor(loginFailureAllowance+2))
	assert.Equal(t, 8*time.Second, loginBackoffFor(loginFailureAllowance+3))

	// And it stops growing, so an account is never locked out permanently.
	assert.Equal(t, loginBackoffCeiling, loginBackoffFor(30))
	assert.Equal(t, loginBackoffCeiling, loginBackoffFor(10000))
}

// newTestThrottle builds a throttle whose clock the test controls.
func newTestThrottle(now *time.Time) *loginThrottle {
	throttle := newLoginThrottle()
	throttle.now = func() time.Time { return *now }
	return throttle
}

func TestLoginThrottle_AllowsTheAllowanceThenBlocks(t *testing.T) {
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	throttle := newTestThrottle(&now)

	for attempt := 1; attempt <= loginFailureAllowance; attempt++ {
		assert.Zero(t, throttle.recordFailure("alice"), "attempt %d must not delay", attempt)
		assert.Zero(t, throttle.retryAfter("alice"), "attempt %d must not block", attempt)
	}

	delay := throttle.recordFailure("alice")
	assert.Equal(t, loginBackoffBase, delay)
	assert.Equal(t, loginBackoffBase, throttle.retryAfter("alice"))

	// The block expires with the clock, not with a restart.
	now = now.Add(loginBackoffBase)
	assert.Zero(t, throttle.retryAfter("alice"))
}

func TestLoginThrottle_BlocksIndependentlyPerAccount(t *testing.T) {
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	throttle := newTestThrottle(&now)

	for attempt := 0; attempt < loginFailureAllowance+1; attempt++ {
		throttle.recordFailure("alice")
	}
	assert.Positive(t, throttle.retryAfter("alice"))
	assert.Zero(t, throttle.retryAfter("bob"), "one account's failures must not block another")
}

// TestLoginThrottle_FoldsCase stops a caller from buying a fresh allowance by
// respelling the same account.
func TestLoginThrottle_FoldsCaseAndWhitespace(t *testing.T) {
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	throttle := newTestThrottle(&now)

	for attempt := 0; attempt < loginFailureAllowance+1; attempt++ {
		throttle.recordFailure("alice")
	}

	assert.Positive(t, throttle.retryAfter("ALICE"))
	assert.Positive(t, throttle.retryAfter("  Alice  "))
}

func TestLoginThrottle_ResetClearsTheRecord(t *testing.T) {
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	throttle := newTestThrottle(&now)

	for attempt := 0; attempt < loginFailureAllowance+1; attempt++ {
		throttle.recordFailure("alice")
	}
	require.Positive(t, throttle.retryAfter("alice"))

	throttle.reset("alice")
	assert.Zero(t, throttle.retryAfter("alice"))

	// A clean slate means the allowance applies again.
	for attempt := 1; attempt <= loginFailureAllowance; attempt++ {
		assert.Zero(t, throttle.recordFailure("alice"))
	}
}

// TestLoginThrottle_ForgetsIdleAccounts keeps the map from growing without bound
// over a long uptime.
func TestLoginThrottle_ForgetsIdleAccounts(t *testing.T) {
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	throttle := newTestThrottle(&now)

	throttle.recordFailure("alice")
	require.Len(t, throttle.records, 1)

	// Past the window, and no longer blocking anyone, the record is collected on
	// the next sweep.
	now = now.Add(loginAttemptWindow + loginSweepInterval)
	throttle.recordFailure("bob")

	assert.NotContains(t, throttle.records, "alice")
	assert.Contains(t, throttle.records, "bob")
}

func TestLoginThrottle_EmptyUsernameIsIgnored(t *testing.T) {
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	throttle := newTestThrottle(&now)

	assert.Zero(t, throttle.recordFailure(""))
	assert.Zero(t, throttle.retryAfter(""))
	assert.Empty(t, throttle.records)
}
