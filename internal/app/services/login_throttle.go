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

	// maxConcurrentLoginVerifications caps how many passwords are being checked at
	// once, across every account and every caller.
	//
	// Per-account counting cannot be the only ceiling: when the account table is
	// saturated with records that are all blocking or in flight, an untracked name
	// is admitted without a reservation, and a caller spreading requests across many
	// client addresses would otherwise reach it with nothing bounding the rate. A
	// process-wide ceiling does not care which name or which address, so it caps
	// guessing in exactly the state where the per-account checks have no opinion.
	// Eight slot holders is far above a console's real concurrent sign-ins and far
	// below what a brute force needs.
	maxConcurrentLoginVerifications = 8
)

// loginAttemptRecord is one account's recent failed sign-ins.
type loginAttemptRecord struct {
	failures     int
	blockedUntil time.Time
	lastFailure  time.Time
	// inFlight is the number of attempts currently verifying a password for this
	// account. It is what stops a burst of concurrent requests from every one of
	// them observing an unblocked account and reaching the credential check.
	inFlight int
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

	// verifications is how many password checks are in progress, across every
	// account and caller. Guarded by mutex.
	verifications int
}

func newLoginThrottle() *loginThrottle {
	return &loginThrottle{
		records: make(map[string]*loginAttemptRecord),
		now:     time.Now,
	}
}

// throttleKey normalises an account name into the throttle's key.
//
// Case is *not* folded. Account lookup is case-sensitive and the column has no
// case-insensitive collation, so "alice" and "ALICE" can be two different
// accounts; sharing one record would let whoever controls the second clear the
// first's failures by signing in. Folding does not help guessing either: a password
// submitted for a respelled name is checked against a different account, so it can
// never be a correct guess for this one.
func throttleKey(username string) string {
	return strings.TrimSpace(username)
}

// admit decides whether an attempt may proceed.
//
// It returns the delay to report when it may not, and reserves the account's
// in-flight slot when it may. A caller that is admitted must call finish exactly
// once, or the account stays occupied.
//
// The reservation is what closes the window a plain check leaves open: a burst of
// concurrent attempts would otherwise all read an unblocked account before any of
// them recorded a failure, and every one of them would reach the password
// comparison.
func (t *loginThrottle) admit(username string) (retryAfter time.Duration, admitted bool) {
	key := throttleKey(username)

	t.mutex.Lock()
	defer t.mutex.Unlock()

	// The process-wide ceiling comes first, and covers every admitted attempt
	// including one this throttle declines to track, so every admission is counted
	// against it and finish always has one to give back.
	if t.verifications >= maxConcurrentLoginVerifications {
		return time.Second, false
	}

	if key == "" {
		t.verifications++
		return 0, true
	}

	now := t.now()
	t.sweepLocked(now)
	record := t.liveRecordLocked(key, now)

	if record != nil {
		if remaining := record.blockedUntil.Sub(now); remaining > 0 {
			return remaining, false
		}
		// One verification at a time per account. A second concurrent attempt is
		// refused instead of being allowed to check the password too, which keeps
		// the number of verifications bounded by the throttle rather than by the
		// caller's parallelism. A double-submitted form retries a second later.
		if record.inFlight > 0 {
			return time.Second, false
		}
	} else {
		record = t.reserveRecordLocked(key, now)
		if record == nil {
			// The table is full of other accounts' live records. Refusing to track
			// this one keeps those records, and it fails open for a name that is not
			// being tracked rather than losing protection for names that are. The
			// attempt still holds a process-wide slot, so it is bounded even though
			// it is not counted against an account.
			t.verifications++
			return 0, true
		}
	}

	t.verifications++
	record.inFlight++
	return 0, true
}

// finish releases the in-flight slot and records the outcome.
// finish records the outcome of an admitted attempt and releases its slot. It must
// be called exactly once per admitted attempt.
func (t *loginThrottle) finish(username string, succeeded bool) {
	key := throttleKey(username)

	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Released after the outcome is recorded, so a replacement attempt cannot start
	// verifying before this one has been counted.
	if t.verifications > 0 {
		t.verifications--
	}

	if key == "" {
		return
	}

	record, ok := t.records[key]
	if !ok {
		return
	}
	if record.inFlight > 0 {
		record.inFlight--
	}

	if succeeded {
		delete(t.records, key)
		return
	}

	now := t.now()
	record.failures++
	record.lastFailure = now

	if record.failures > loginFailureAllowance {
		record.blockedUntil = now.Add(loginBackoffFor(record.failures))
	}
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

	record := t.liveRecordLocked(key, t.now())
	if record == nil {
		return 0
	}

	remaining := record.blockedUntil.Sub(t.now())
	if remaining <= 0 {
		return 0
	}
	return remaining
}

// liveRecordLocked returns the account's live record, discarding one that has gone
// a full window without a failure.
//
// The periodic sweep alone is not enough: it runs at most once a minute, so an
// account whose record has aged out can still be found here with its old failure
// count, and the next failure would then be counted on top of an expired history
// and block somebody who has a clean slate. The caller holds the mutex.
func (t *loginThrottle) liveRecordLocked(key string, now time.Time) *loginAttemptRecord {
	record, ok := t.records[key]
	if !ok {
		return nil
	}

	idle := now.Sub(record.lastFailure) > loginAttemptWindow
	blocking := now.Before(record.blockedUntil)
	// A record that is holding a verification is never expired, for the same reason
	// the sweep skips it: dropping it would let a second attempt reserve a
	// replacement record, and the first attempt's outcome would then land on that
	// replacement.
	if idle && !blocking && record.inFlight == 0 {
		delete(t.records, key)
		return nil
	}
	return record
}

// reserveRecordLocked makes room for one more account and returns its record, or
// nil when there is no room to make without discarding a live record. The caller
// holds the mutex.
func (t *loginThrottle) reserveRecordLocked(key string, now time.Time) *loginAttemptRecord {
	if len(t.records) >= maxTrackedLoginAccounts {
		if !t.evictOneLocked(now) {
			return nil
		}
	}

	// lastFailure is stamped at admission, not only when the attempt fails: the
	// record has to look live to the expiry check for as long as the verification
	// it is holding a slot for is still running.
	record := &loginAttemptRecord{lastFailure: now}
	t.records[key] = record
	return record
}

// evictOneLocked frees a single slot by dropping the least recently failed account
// that is not currently blocking anyone, and reports whether it found one.
//
// A record that is blocking is never chosen: discarding it would be a way to clear
// an account's backoff just by filling the table, which is the whole point of the
// table.
func (t *loginThrottle) evictOneLocked(now time.Time) bool {
	var oldestKey string
	var oldestFailure time.Time

	for key, record := range t.records {
		if now.Before(record.blockedUntil) || record.inFlight > 0 {
			continue
		}
		if oldestKey == "" || record.lastFailure.Before(oldestFailure) {
			oldestKey, oldestFailure = key, record.lastFailure
		}
	}

	if oldestKey == "" {
		return false
	}
	delete(t.records, oldestKey)
	return true
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

// sweepLocked drops records that are past the window and no longer holding a
// verification or blocking anyone. The caller holds the mutex.
func (t *loginThrottle) sweepLocked(now time.Time) {
	if now.Sub(t.lastSweep) < loginSweepInterval {
		return
	}
	t.lastSweep = now

	for key, record := range t.records {
		if record.inFlight > 0 || now.Before(record.blockedUntil) {
			continue
		}
		if now.Sub(record.lastFailure) > loginAttemptWindow {
			delete(t.records, key)
		}
	}
}
