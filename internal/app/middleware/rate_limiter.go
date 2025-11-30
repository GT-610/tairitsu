package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"go.uber.org/zap"
)

// TokenBucket represents a token bucket for rate limiting
type TokenBucket struct {
	capacity     int           // Bucket capacity
	tokens       int           // Current token count
	refillRate   int           // Tokens added per second
	lastRefill   time.Time     // Last time tokens were added
	refillMutex  sync.Mutex    // Mutex for thread-safe token refilling
}

// NewTokenBucket creates a new token bucket
func NewTokenBucket(capacity, refillRate int) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,     // Initially full
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// GetToken attempts to get a token from the bucket
// Returns true if successful, false otherwise
func (tb *TokenBucket) GetToken() bool {
	tb.refillMutex.Lock()
	defer tb.refillMutex.Unlock()

	// Refill tokens based on elapsed time
	now := time.Now()
	duration := now.Sub(tb.lastRefill)
	tokensToAdd := int(duration.Seconds()) * tb.refillRate

	if tokensToAdd > 0 {
		// Update token count, not exceeding capacity
		newTokens := tb.tokens + tokensToAdd
		if newTokens > tb.capacity {
			newTokens = tb.capacity
		}
		tb.tokens = newTokens
		tb.lastRefill = now
	}

	// Try to get a token
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// RateLimiter manages rate limiting for multiple clients
type RateLimiter struct {
	buckets     map[string]*TokenBucket // Map of IP addresses to token buckets
	bucketMutex sync.RWMutex            // Mutex for thread-safe bucket access
	capacity    int                     // Default bucket capacity
	refillRate  int                     // Default tokens added per second
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(capacity, refillRate int) *RateLimiter {
	return &RateLimiter{
		buckets:    make(map[string]*TokenBucket),
		capacity:   capacity,
		refillRate: refillRate,
	}
}

// GetBucket gets or creates a token bucket for a client IP
func (rl *RateLimiter) GetBucket(ip string) *TokenBucket {
	rl.bucketMutex.RLock()
	bucket, exists := rl.buckets[ip]
	rl.bucketMutex.RUnlock()

	if exists {
		return bucket
	}

	// Create new token bucket
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
var DefaultRateLimiter = NewRateLimiter(100, 10) // Default: 100 tokens, 10 tokens per second

// RateLimit is the API rate limiting middleware
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP
		clientIP := c.ClientIP()

		// Get or create token bucket for the client
		bucket := DefaultRateLimiter.GetBucket(clientIP)

		// Try to get a token
		if !bucket.GetToken() {
			// Log rate limit event (if logger is initialized)
			defer func() {
				if r := recover(); r != nil {
					// Ignore logger initialization failures
				}
			}()
			logger.Warn("API rate limit triggered", zap.String("client_ip", clientIP), zap.String("path", c.Request.URL.Path))

			// Return 429 Too Many Requests
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests, please try again later",
			})
			c.Abort()
			return
		}

		// Continue processing the request
		c.Next()
	}
}
