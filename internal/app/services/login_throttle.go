package services

import (
	"strings"
	"sync"
	"time"
)

// Sign-in throttling constants.
//
// The transport-level limiter keys on the client address, and a proxy that
// forwards a caller-supplied X-Forwarded-For hands that key to the caller, who
// can then buy a fresh budget per request. The account name is the part of a
// sign-in an attacker cannot vary while attacking a particular account, so it is
// what this counts.
const (
	// loginFailureAllowance is how many recent failures an account may collect
	// before any delay applies. People mistype passwords, and a handful of tries
	// is not an attack.
	loginFailureAllowance = 5

	// loginBackoffBase is the delay the first failure past the allowance earns.
	// Each further failure doubles it.
	loginBackoffBase = 2 * time.Second

	// loginBackoffCeiling caps the delay, so an account that is being hammered
	// still becomes usable again on its own rather than staying shut until the
	// process restarts.
	loginBackoffCeiling = 5 * time.Minute

	// loginAttemptWindow is how long a record survives without a new failure.
	// Past it the account starts from a clean slate.
	loginAttemptWindow = 30 * time.Minute

	// maxTrackedLoginAccounts bounds the memory the throttle can hold.
	maxTrackedLoginAccounts = 10000

	// loginSweepInterval limits how often expired records are collected, so a
	// failed sign-in does not pay for a full scan.
	loginSweepInterval = time.Minute
)

// loginAttemptRecord is one account's recent failed sign-ins.
type loginAttemptRecord struct {
	failures     int
	blockedUntil time.Time
	lastFailure  time.Time
}

// loginThrottle counts failed sign-ins per account and answers how long a caller
// has to wait before the next attempt is examined at all.
//
// It is per UserService rather than package-level so that a test starts from a
// clean slate without the order the tests happen to run in mattering.
type loginThrottle struct {
	mutex     sync.Mutex
	records   map[string]*loginAttemptRecord
	lastSweep time.Time
	now       func() time.Time
}

func newLoginThrottle() *loginThrottle {
	return &loginThrottle{
		records: make(map[string]*loginAttemptRecord),
		now:     time.Now,
	}
}

// throttleKey normalises an account name into the throttle's key.
//
// Case is folded because account lookup is exact-match while a caller is free to
// vary the spelling they try; without folding, "Alice" and "alice" would count
// as different targets even though a human reads them as the same account. This
// only ever makes the throttle stricter, and never gives an attacker a way to
// reach an account they could not already reach.
func throttleKey(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

// retryAfter reports how long the caller must wait before another attempt at this
// account will be examined. Zero means it may proceed now.
func (t *loginThrottle) retryAfter(username string) time.Duration {
	key := throttleKey(username)
	if key == "" {
		return 0
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	record, ok := t.records[key]
	if !ok {
		return 0
	}

	remaining := record.blockedUntil.Sub(t.now())
	if remaining <= 0 {
		return 0
	}
	return remaining
}

// recordFailure counts one failed attempt and reports the delay it earned.
func (t *loginThrottle) recordFailure(username string) time.Duration {
	key := throttleKey(username)
	if key == "" {
		return 0
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	now := t.now()
	t.sweepLocked(now)

	record, ok := t.records[key]
	if !ok {
		if len(t.records) >= maxTrackedLoginAccounts {
			// Every slot is taken by other accounts. Forgetting them costs one
			// round of forgiveness, which is better than letting the map grow
			// until the process runs out of memory.
			t.records = make(map[string]*loginAttemptRecord)
		}
		record = &loginAttemptRecord{}
		t.records[key] = record
	}

	record.failures++
	record.lastFailure = now

	if record.failures <= loginFailureAllowance {
		return 0
	}

	delay := loginBackoffFor(record.failures)
	record.blockedUntil = now.Add(delay)
	return delay
}

// reset forgets an account's failures after a successful sign-in.
func (t *loginThrottle) reset(username string) {
	key := throttleKey(username)
	if key == "" {
		return
	}

	t.mutex.Lock()
	delete(t.records, key)
	t.mutex.Unlock()
}

// loginBackoffFor returns the delay a failure count earns: nothing up to the
// allowance, then doubling from loginBackoffBase up to loginBackoffCeiling.
func loginBackoffFor(failures int) time.Duration {
	if failures <= loginFailureAllowance {
		return 0
	}

	delay := loginBackoffBase
	for attempt := loginFailureAllowance + 2; attempt <= failures; attempt++ {
		if delay >= loginBackoffCeiling {
			return loginBackoffCeiling
		}
		delay *= 2
	}

	if delay > loginBackoffCeiling {
		return loginBackoffCeiling
	}
	return delay
}

// sweepLocked drops records that are past the window and no longer blocking
// anyone. The caller holds the mutex.
func (t *loginThrottle) sweepLocked(now time.Time) {
	if now.Sub(t.lastSweep) < loginSweepInterval {
		return
	}
	t.lastSweep = now

	for key, record := range t.records {
		idle := now.Sub(record.lastFailure) > loginAttemptWindow
		if idle && !now.Before(record.blockedUntil) {
			delete(t.records, key)
		}
	}
}
