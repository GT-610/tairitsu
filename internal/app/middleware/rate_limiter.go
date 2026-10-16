package middleware

import (
	"strings"
	"sync"
	"time"

	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// Rate limiting uses the token bucket algorithm, keyed by client address.

// tokenBucket holds the remaining budget for one client.
type tokenBucket struct {
	capacity   int       // Bucket capacity
	tokens     int       // Current number of tokens
	refillRate int       // Tokens added per second
	lastRefill time.Time // Last time tokens were refilled
	mutex      sync.Mutex
	now        func() time.Time
}

// newTokenBucket creates a new token bucket that reads the wall clock.
func newTokenBucket(capacity, refillRate int) *tokenBucket {
	return newTokenBucketWithClock(capacity, refillRate, time.Now)
}

// newTokenBucketWithClock creates a token bucket that reads the time from now,
// so that refill behaviour can be tested without waiting for it.
func newTokenBucketWithClock(capacity, refillRate int, now func() time.Time) *tokenBucket {
	return &tokenBucket{
		capacity:   capacity,
		tokens:     capacity, // Initially full
		refillRate: refillRate,
		lastRefill: now(),
		now:        now,
	}
}

// getToken attempts to acquire a token, reporting whether it succeeded.
func (tb *tokenBucket) getToken() bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	// Refill tokens
	now := tb.now()
	duration := now.Sub(tb.lastRefill)
	tokensToAdd := int(duration.Seconds()) * tb.refillRate

	if tokensToAdd > 0 {
		newTokens := tb.tokens + tokensToAdd
		if newTokens > tb.capacity {
			newTokens = tb.capacity
		}
		tb.tokens = newTokens
		tb.lastRefill = now
	}

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// rateLimiter tracks one bucket per client address.
type rateLimiter struct {
	buckets     map[string]*tokenBucket // Client address to token bucket mapping
	bucketMutex sync.RWMutex            // Protects buckets
	capacity    int                     // Default bucket capacity
	refillRate  int                     // Default tokens added per second
	maxBuckets  int                     // Maximum number of tracked clients
	denyBucket  *tokenBucket            // Reusable zero-token bucket for over-capacity clients
	now         func() time.Time
}

// maxTrackedClients bounds the memory a rate limiter can hold.
const maxTrackedClients = 10000

const cleanupInterval = 5 * time.Minute
const bucketStaleThreshold = 10 * time.Minute

// newRateLimiter creates a rate limiter that reads the wall clock.
func newRateLimiter(capacity, refillRate int) *rateLimiter {
	return newRateLimiterWithClock(capacity, refillRate, time.Now)
}

func newRateLimiterWithClock(capacity, refillRate int, now func() time.Time) *rateLimiter {
	return &rateLimiter{
		buckets:    make(map[string]*tokenBucket),
		capacity:   capacity,
		refillRate: refillRate,
		maxBuckets: maxTrackedClients,
		denyBucket: newTokenBucketWithClock(0, 0, now),
		now:        now,
	}
}

// startCleanup drops buckets that no longer matter, so that a long-running
// process does not keep one per address it has ever seen. Only the limiters
// that live for the whole process need it.
func (rl *rateLimiter) startCleanup() {
	go rl.cleanupLoop()
}

func (rl *rateLimiter) cleanupLoop() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		rl.cleanupStale()
	}
}

func (rl *rateLimiter) cleanupStale() {
	rl.bucketMutex.Lock()
	defer rl.bucketMutex.Unlock()

	now := rl.now()
	for address, bucket := range rl.buckets {
		bucket.mutex.Lock()
		stale := now.Sub(bucket.lastRefill) > bucketStaleThreshold
		bucket.mutex.Unlock()

		if stale {
			delete(rl.buckets, address)
		}
	}
}

// getBucket retrieves or creates the token bucket for the given client address.
func (rl *rateLimiter) getBucket(address string) *tokenBucket {
	rl.bucketMutex.RLock()
	bucket, exists := rl.buckets[address]
	rl.bucketMutex.RUnlock()

	if exists {
		return bucket
	}

	newBucket := newTokenBucketWithClock(rl.capacity, rl.refillRate, rl.now)

	rl.bucketMutex.Lock()
	defer rl.bucketMutex.Unlock()

	// Another goroutine may have inserted this address while the lock was free.
	if bucket, exists := rl.buckets[address]; exists {
		return bucket
	}

	// Reject new client addresses once the map is at capacity, so a flood of
	// distinct addresses cannot grow it without bound.
	if len(rl.buckets) >= rl.maxBuckets {
		return rl.denyBucket
	}

	rl.buckets[address] = newBucket
	return newBucket
}

// apiRateLimiter is the budget for the JSON surface.
var apiRateLimiter = newRateLimiter(100, 10) // 100 tokens, refills 10 per second

// authenticationRateLimiter is the stricter budget for authentication endpoints.
var authenticationRateLimiter = newRateLimiter(20, 2) // 20 tokens, refills 2 per second

func init() {
	apiRateLimiter.startCleanup()
	authenticationRateLimiter.startCleanup()
}

// RateLimit is the API rate limiting middleware.
func RateLimit() fiber.Handler {
	return rateLimitHandler(apiRateLimiter)
}

// AuthRateLimit is the stricter rate limiting middleware for auth endpoints.
func AuthRateLimit() fiber.Handler {
	return rateLimitHandler(authenticationRateLimiter)
}

func rateLimitHandler(limiter *rateLimiter) fiber.Handler {
	return func(c fiber.Ctx) error {
		// The context is only valid for this request, so the bucket map must not
		// hold a reference into its buffer.
		clientAddress := strings.Clone(c.IP())
		bucket := limiter.getBucket(clientAddress)

		if !bucket.getToken() {
			logger.Warn("API rate limit triggered", zap.String("client_ip", clientAddress), zap.String("path", c.Path()))

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":      "Too many requests. Please try again later.",
				"message":    "Too many requests. Please try again later.",
				"error_code": "system.rate_limited",
			})
		}

		return c.Next()
	}
}
