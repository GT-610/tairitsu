package services

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoginBackoffGrowsThenCaps pins the shape of the delay: nothing up to the
// allowance, then doubling, then a ceiling so an account recovers on its own.
func TestLoginBackoffGrowsThenCaps(t *testing.T) {
	for failures := 0; failures <= loginFailureAllowance; failures++ {
		assert.Zero(t, loginBackoffFor(failures), "failures=%d", failures)
	}

	assert.Equal(t, 2*time.Second, loginBackoffFor(loginFailureAllowance+1))
	assert.Equal(t, 4*time.Second, loginBackoffFor(loginFailureAllowance+2))
	assert.Equal(t, 8*time.Second, loginBackoffFor(loginFailureAllowance+3))

	assert.Equal(t, loginBackoffCeiling, loginBackoffFor(30))
	assert.Equal(t, loginBackoffCeiling, loginBackoffFor(10000))
}

// newTestThrottle builds a throttle whose clock the test controls.
func newTestThrottle(now *time.Time) *loginThrottle {
	throttle := newLoginThrottle()
	throttle.now = func() time.Time { return *now }
	return throttle
}

// spendFailures admits and fails one attempt at a time, the way a client retrying
// a wrong password does.
func spendFailures(t *testing.T, throttle *loginThrottle, username string, attempts int) {
	t.Helper()
	for attempt := 0; attempt < attempts; attempt++ {
		retryAfter, admitted := throttle.admit(username)
		require.True(t, admitted, "attempt %d was refused with %v", attempt, retryAfter)
		throttle.finish(username, false)
	}
}

func TestLoginThrottle_AllowsTheAllowanceThenBlocks(t *testing.T) {
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	throttle := newTestThrottle(&now)

	for attempt := 1; attempt <= loginFailureAllowance; attempt++ {
		retryAfter, admitted := throttle.admit("alice")
		require.True(t, admitted, "attempt %d must be allowed", attempt)
		assert.Zero(t, retryAfter)
		throttle.finish("alice", false)
	}

	// The failure past the allowance blocks the next admission.
	throttle.finish("alice", false)
	retryAfter, admitted := throttle.admit("alice")
	assert.False(t, admitted)
	assert.Positive(t, retryAfter)

	// The block expires with the clock, not with a restart.
	now = now.Add(loginBackoffCeiling)
	_, admitted = throttle.admit("alice")
	assert.True(t, admitted)
}

func TestLoginThrottle_BlocksIndependentlyPerAccount(t *testing.T) {
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	throttle := newTestThrottle(&now)

	spendFailures(t, throttle, "alice", loginFailureAllowance+1)

	_, admitted := throttle.admit("alice")
	assert.False(t, admitted)
	_, admitted = throttle.admit("bob")
	assert.True(t, admitted, "one account's failures must not block another")
}

// TestLoginThrottle_KeysOnTheExactUsername is the regression test for sharing one
// record between two accounts whose names differ only by case. Account lookup is
// case-sensitive, so they can both exist, and a shared record let whoever holds
// the second clear the first's failures by signing in.
func TestLoginThrottle_KeysOnTheExactUsername(t *testing.T) {
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	throttle := newTestThrottle(&now)

	spendFailures(t, throttle, "alice", loginFailureAllowance+1)
	_, admitted := throttle.admit("alice")
	require.False(t, admitted, "precondition: alice is blocked")

	// A successful sign-in to the other account must not clear alice's record.
	retryAfter, admitted := throttle.admit("ALICE")
	require.True(t, admitted, "a different account is not blocked: %v", retryAfter)
	throttle.finish("ALICE", true)

	_, admitted = throttle.admit("alice")
	assert.False(t, admitted, "the other account's success must not clear alice's block")
}

func TestLoginThrottle_TrimsWhitespace(t *testing.T) {
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	throttle := newTestThrottle(&now)

	spendFailures(t, throttle, "alice", loginFailureAllowance+1)

	_, admitted := throttle.admit("  alice  ")
	assert.False(t, admitted, "padding must not buy a fresh allowance")
}

// TestLoginThrottle_ForgetsAnExpiredRecordBeforeCountingOnIt is the regression test
// for resuming an aged-out count.
//
// The periodic sweep is time-gated, so it can be skipped while a record is already
// past the window: the record ages out between two sweeps that are less than the
// sweep interval apart. Counting on top of it would block an account that has a
// clean slate, so the record's own age has to be checked when it is read.
func TestLoginThrottle_ForgetsAnExpiredRecordBeforeCountingOnIt(t *testing.T) {
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	throttle := newTestThrottle(&now)

	// Use the allowance without blocking, so the record is idle but not blocking.
	spendFailures(t, throttle, "alice", loginFailureAllowance)

	// A sweep runs just before the record ages out, which keeps it and moves the
	// sweep clock forward.
	now = now.Add(loginAttemptWindow - 24*time.Second)
	spendFailures(t, throttle, "sweeper", 1)
	require.Contains(t, throttle.records, "alice")

	// The record now ages out, but less than the sweep interval has passed, so the
	// next lookup will not sweep.
	now = now.Add(36 * time.Second)
	require.True(t, now.Sub(recordLastFailure(throttle, "alice")) > loginAttemptWindow,
		"precondition: alice's record is past the window")

	// One more failure must start a clean count rather than adding to the expired
	// one, which would block her.
	spendFailures(t, throttle, "alice", 1)

	_, admitted := throttle.admit("alice")
	assert.True(t, admitted, "an expired history must not count against a fresh attempt")
	if record, ok := throttle.records["alice"]; ok {
		assert.Equal(t, 1, record.failures, "the count must have restarted")
	}
}

// recordLastFailure reads the recorded failure time for an account.
func recordLastFailure(throttle *loginThrottle, username string) time.Time {
	throttle.mutex.Lock()
	defer throttle.mutex.Unlock()
	if record, ok := throttle.records[throttleKey(username)]; ok {
		return record.lastFailure
	}
	return time.Time{}
}

// TestLoginThrottle_CapacityDoesNotDiscardBlocks is the regression test for clearing
// the table to make room: an unauthenticated caller could fill it with distinct
// names and drop the target account's block without waiting for Retry-After.
func TestLoginThrottle_CapacityDoesNotDiscardBlocks(t *testing.T) {
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	throttle := newTestThrottle(&now)

	// Block the account we care about.
	spendFailures(t, throttle, "target", loginFailureAllowance+1)
	_, admitted := throttle.admit("target")
	require.False(t, admitted, "precondition: the target is blocked")

	// Fill every other slot with names that are not blocking anyone.
	for index := 0; index < maxTrackedLoginAccounts; index++ {
		spendFailures(t, throttle, fmt.Sprintf("filler-%d", index), 1)
	}
	require.Len(t, throttle.records, maxTrackedLoginAccounts)

	// One more failure forces an eviction.
	spendFailures(t, throttle, "latecomer", 1)

	_, admitted = throttle.admit("target")
	assert.False(t, admitted, "filling the table must not clear a block")
}

// TestLoginThrottle_RefusesConcurrentVerifications is the regression test for a
// burst of parallel attempts: every one of them read an unblocked account before
// any had recorded a failure, so all of them reached the password comparison.
func TestLoginThrottle_RefusesConcurrentVerifications(t *testing.T) {
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	throttle := newTestThrottle(&now)

	_, admitted := throttle.admit("alice")
	require.True(t, admitted, "the first attempt proceeds")

	// While it is still verifying, a second does not get its own check.
	retryAfter, admitted := throttle.admit("alice")
	assert.False(t, admitted)
	assert.Positive(t, retryAfter)

	throttle.finish("alice", false)

	// Once it completes, the next attempt is admitted again.
	_, admitted = throttle.admit("alice")
	assert.True(t, admitted)
}

// TestLoginThrottle_BoundsConcurrentVerificationsUnderLoad is the same property
// observed the way an attacker would apply it: many goroutines at once.
//
// Every worker attempts admission before the accepted one is released, so the count
// cannot drift because a late worker legitimately found the slot free again.
func TestLoginThrottle_BoundsConcurrentVerificationsUnderLoad(t *testing.T) {
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	throttle := newTestThrottle(&now)

	const workers = 50
	var attempted, finished sync.WaitGroup
	var accepted int32
	release := make(chan struct{})

	attempted.Add(workers)
	finished.Add(workers)
	for worker := 0; worker < workers; worker++ {
		go func() {
			defer finished.Done()

			_, admitted := throttle.admit("alice")
			attempted.Done()

			if !admitted {
				return
			}
			atomic.AddInt32(&accepted, 1)
			// Hold the verification until every worker has tried, so the assertion
			// below cannot be disturbed by one arriving after the slot was freed.
			<-release
			throttle.finish("alice", false)
		}()
	}

	attempted.Wait()
	close(release)
	finished.Wait()

	assert.Equal(t, int32(1), accepted, "only one verification may be in flight per account")
}

// TestLoginThrottle_BoundsAnUntrackedAccount is the regression test for the
// fail-open path. When the table is saturated with records that must not be
// discarded, an account that is not tracked is admitted without a reservation — so
// per-account counting has no opinion about it. The process-wide ceiling still
// applies, which is what bounds a caller spreading requests across client addresses.
func TestLoginThrottle_BoundsAnUntrackedAccount(t *testing.T) {
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	throttle := newTestThrottle(&now)

	// Fill every slot with an account that is currently blocked, so nothing is
	// evictable and the next new name falls through to the fail-open path.
	for index := 0; index < maxTrackedLoginAccounts; index++ {
		spendFailures(t, throttle, fmt.Sprintf("blocked-%d", index), loginFailureAllowance+1)
	}
	require.Len(t, throttle.records, maxTrackedLoginAccounts)

	// The untracked name is admitted, because there is no room to track it...
	held := make([]string, 0, maxConcurrentLoginVerifications)
	for index := 0; index < maxConcurrentLoginVerifications; index++ {
		name := fmt.Sprintf("untracked-%d", index)
		_, admitted := throttle.admit(name)
		require.True(t, admitted, "the fail-open path admits %s", name)
		require.NotContains(t, throttle.records, name, "precondition: it is not being tracked")
		held = append(held, name)
	}

	// ...but not without limit: the process-wide ceiling still holds.
	_, admitted := throttle.admit("untracked-overflow")
	assert.False(t, admitted, "an untracked account is still bounded process-wide")

	for _, name := range held {
		throttle.finish(name, false)
	}
}

// TestLoginThrottle_DoesNotExpireARecordHoldingAVerification covers the case the
// sweep already guards: a verification that outlives the window must not have its
// record discarded, or a second attempt could reserve a replacement and the first
// attempt's outcome would land on it.
func TestLoginThrottle_DoesNotExpireARecordHoldingAVerification(t *testing.T) {
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	throttle := newTestThrottle(&now)

	_, admitted := throttle.admit("alice")
	require.True(t, admitted)

	// The verification takes longer than the expiry window allows for.
	now = now.Add(loginAttemptWindow + time.Minute)

	retryAfter, admitted := throttle.admit("alice")
	assert.False(t, admitted, "the record is still holding a verification")
	assert.Positive(t, retryAfter)

	throttle.finish("alice", false)

	record, ok := throttle.records["alice"]
	require.True(t, ok, "the outcome must land on the record it belongs to")
	assert.Equal(t, 1, record.failures)
	assert.Equal(t, 0, record.inFlight)
}

func TestLoginThrottle_ForgetsIdleAccounts(t *testing.T) {
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	throttle := newTestThrottle(&now)

	spendFailures(t, throttle, "alice", 1)
	require.Len(t, throttle.records, 1)

	now = now.Add(loginAttemptWindow + loginSweepInterval)
	spendFailures(t, throttle, "bob", 1)

	assert.NotContains(t, throttle.records, "alice")
	assert.Contains(t, throttle.records, "bob")
}

func TestLoginThrottle_EmptyUsernameIsIgnored(t *testing.T) {
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	throttle := newTestThrottle(&now)

	_, admitted := throttle.admit("")
	assert.True(t, admitted)
	throttle.finish("", false)
	assert.Empty(t, throttle.records)
}
