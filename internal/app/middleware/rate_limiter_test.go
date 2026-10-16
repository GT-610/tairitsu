package middleware

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// fakeClock lets the refill behaviour be tested without waiting for it.
type fakeClock struct {
	current time.Time
}

func newFakeClock() *fakeClock {
	return &fakeClock{current: time.Unix(1700000000, 0)}
}

func (c *fakeClock) now() time.Time {
	return c.current
}

func (c *fakeClock) advance(duration time.Duration) {
	c.current = c.current.Add(duration)
}

func TestTokenBucketSpendsAndRefillsWithoutWaiting(t *testing.T) {
	clock := newFakeClock()
	bucket := newTokenBucketWithClock(5, 1, clock.now)

	for remaining := 0; remaining < 5; remaining++ {
		assert.True(t, bucket.getToken())
	}
	assert.False(t, bucket.getToken(), "the bucket must be empty after its capacity is spent")

	clock.advance(500 * time.Millisecond)
	assert.False(t, bucket.getToken(), "a partial second does not refill")

	// The elapsed fraction is kept because lastRefill only moves when a whole
	// token has been added.
	clock.advance(500 * time.Millisecond)
	assert.True(t, bucket.getToken(), "one second refills one token")
	assert.False(t, bucket.getToken())
}

func TestTokenBucketRefillIsCappedAtCapacity(t *testing.T) {
	clock := newFakeClock()
	bucket := newTokenBucketWithClock(3, 10, clock.now)

	assert.True(t, bucket.getToken())
	clock.advance(time.Hour)

	for remaining := 0; remaining < 3; remaining++ {
		assert.True(t, bucket.getToken())
	}
	assert.False(t, bucket.getToken(), "an idle bucket must not accumulate beyond its capacity")
}

func TestRateLimiterReturnsTheSameBucketForOneAddress(t *testing.T) {
	limiter := newRateLimiterWithClock(10, 2, newFakeClock().now)

	first := limiter.getBucket("192.0.2.1")
	second := limiter.getBucket("192.0.2.2")
	again := limiter.getBucket("192.0.2.1")

	assert.NotNil(t, first)
	assert.NotNil(t, second)
	assert.NotSame(t, first, second)
	assert.Same(t, first, again)
}

// An unbounded map keyed by client address is a memory exhaustion vector, so
// new addresses are refused once the map is full and share one empty bucket.
func TestRateLimiterRefusesNewAddressesAtCapacity(t *testing.T) {
	limiter := newRateLimiterWithClock(5, 1, newFakeClock().now)
	limiter.maxBuckets = 2

	first := limiter.getBucket("192.0.2.1")
	second := limiter.getBucket("192.0.2.2")
	overflow := limiter.getBucket("192.0.2.3")

	assert.NotSame(t, first, overflow)
	assert.NotSame(t, second, overflow)
	assert.Same(t, limiter.denyBucket, overflow)
	assert.False(t, overflow.getToken(), "an address over the cap must not be served")
	assert.Len(t, limiter.buckets, 2, "the map must not grow past its cap")

	// An address that was already tracked keeps working.
	assert.NotSame(t, limiter.denyBucket, limiter.getBucket("192.0.2.1"))
}

func TestRateLimiterCleanupDropsIdleBuckets(t *testing.T) {
	clock := newFakeClock()
	limiter := newRateLimiterWithClock(5, 1, clock.now)

	limiter.getBucket("192.0.2.1")
	limiter.getBucket("192.0.2.2")
	assert.Len(t, limiter.buckets, 2)

	clock.advance(bucketStaleThreshold + time.Minute)

	// Using a bucket refreshes it, which is what marks it as still interesting.
	assert.True(t, limiter.getBucket("192.0.2.1").getToken())
	limiter.cleanupStale()

	assert.Len(t, limiter.buckets, 1, "the bucket that was not used must be dropped")
	assert.Contains(t, limiter.buckets, "192.0.2.1")
}
