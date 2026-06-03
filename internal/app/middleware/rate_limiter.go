package middleware

import (
	"sync"
	"time"

	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// RateLimiter middleware for API rate limiting.
// Uses the token bucket algorithm.

// TokenBucket represents a token bucket structure
type TokenBucket struct {
	capacity    int        // Bucket capacity
	tokens      int        // Current number of tokens
	refillRate  int        // Tokens added per second
	lastRefill  time.Time  // Last time tokens were refilled
	refillMutex sync.Mutex // Mutex for refilling tokens
}

// NewTokenBucket creates a new token bucket
func NewTokenBucket(capacity, refillRate int) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity, // Initially full
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// GetToken attempts to acquire a token.
// Returns true on success, false on failure.
func (tb *TokenBucket) GetToken() bool {
	tb.refillMutex.Lock()
	defer tb.refillMutex.Unlock()

	// Refill tokens
	now := time.Now()
	duration := now.Sub(tb.lastRefill)
	tokensToAdd := int(duration.Seconds()) * tb.refillRate

	if tokensToAdd > 0 {
		// Update token count, capped at capacity
		newTokens := tb.tokens + tokensToAdd
		if newTokens > tb.capacity {
			newTokens = tb.capacity
		}
		tb.tokens = newTokens
		tb.lastRefill = now
	}

	// Attempt to acquire a token
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// RateLimiter is the rate limiter
type RateLimiter struct {
	buckets     map[string]*TokenBucket // IP address to token bucket mapping
	bucketMutex sync.RWMutex            // Mutex for accessing buckets
	capacity    int                     // Default bucket capacity
	refillRate  int                     // Default tokens added per second
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(capacity, refillRate int) *RateLimiter {
	rl := &RateLimiter{
		buckets:    make(map[string]*TokenBucket),
		capacity:   capacity,
		refillRate: refillRate,
	}
	go rl.cleanupLoop()
	return rl
}

const cleanupInterval = 5 * time.Minute
const bucketStaleThreshold = 10 * time.Minute

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		rl.cleanupStale()
	}
}

func (rl *RateLimiter) cleanupStale() {
	rl.bucketMutex.Lock()
	defer rl.bucketMutex.Unlock()
	now := time.Now()
	for ip, bucket := range rl.buckets {
		bucket.refillMutex.Lock()
		stale := now.Sub(bucket.lastRefill) > bucketStaleThreshold
		bucket.refillMutex.Unlock()
		if stale {
			delete(rl.buckets, ip)
		}
	}
}

// GetBucket retrieves or creates the token bucket for the given IP
func (rl *RateLimiter) GetBucket(ip string) *TokenBucket {
	rl.bucketMutex.RLock()
	bucket, exists := rl.buckets[ip]
	rl.bucketMutex.RUnlock()

	if exists {
		return bucket
	}

	// Create a new token bucket
	newBucket := NewTokenBucket(rl.capacity, rl.refillRate)

	rl.bucketMutex.Lock()
	// Double-check to prevent race conditions
	if bucket, exists := rl.buckets[ip]; exists {
		rl.bucketMutex.Unlock()
		return bucket
	}

	rl.buckets[ip] = newBucket
	rl.bucketMutex.Unlock()

	return newBucket
}

// DefaultRateLimiter is the default rate limiter instance
var DefaultRateLimiter = NewRateLimiter(100, 10) // 100 tokens, refills 10 per second

// RateLimit is the API rate limiting middleware
func RateLimit() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Get client IP
		clientIP := c.IP()

		// Get or create token bucket
		bucket := DefaultRateLimiter.GetBucket(clientIP)

		// Attempt to acquire a token
		if !bucket.GetToken() {
			logger.Warn("API rate limit triggered", zap.String("client_ip", clientIP), zap.String("path", c.Path()))

			// Return 429 Too Many Requests
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":      "Too many requests. Please try again later.",
				"message":    "Too many requests. Please try again later.",
				"error_code": "system.rate_limited",
				"code":       fiber.StatusTooManyRequests,
			})
		}

		// Continue processing the request
		return c.Next()
	}
}
