package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/GT-610/tairitsu/internal/app/logger"
	"go.uber.org/zap"
)

// RateLimiter 速率限制中间件
// 使用令牌桶算法实现API速率限制

// TokenBucket 令牌桶结构
type TokenBucket struct {
	capacity     int           // 桶容量
	tokens       int           // 当前令牌数
	refillRate   int           // 每秒补充令牌数
	lastRefill   time.Time     // 上次补充令牌时间
	refillMutex  sync.Mutex    // 补充令牌的互斥锁
}

// NewTokenBucket 创建新的令牌桶
func NewTokenBucket(capacity, refillRate int) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,     // 初始填满令牌
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// GetToken 尝试获取一个令牌
// 返回true表示获取成功，false表示失败
func (tb *TokenBucket) GetToken() bool {
	tb.refillMutex.Lock()
	defer tb.refillMutex.Unlock()

	// 补充令牌
	now := time.Now()
	duration := now.Sub(tb.lastRefill)
	tokensToAdd := int(duration.Seconds()) * tb.refillRate

	if tokensToAdd > 0 {
			// 更新令牌数，不超过容量
			newTokens := tb.tokens + tokensToAdd
			if newTokens > tb.capacity {
				newTokens = tb.capacity
			}
			tb.tokens = newTokens
			tb.lastRefill = now
		}

	// 尝试获取令牌
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// RateLimiter 速率限制器
type RateLimiter struct {
	buckets     map[string]*TokenBucket // IP地址到令牌桶的映射
	bucketMutex sync.RWMutex            // 访问buckets的互斥锁
	capacity    int                     // 默认桶容量
	refillRate  int                     // 默认每秒补充令牌数
}

// NewRateLimiter 创建新的速率限制器
func NewRateLimiter(capacity, refillRate int) *RateLimiter {
	return &RateLimiter{
		buckets:    make(map[string]*TokenBucket),
		capacity:   capacity,
		refillRate: refillRate,
	}
}

// GetBucket 获取或创建IP对应的令牌桶
func (rl *RateLimiter) GetBucket(ip string) *TokenBucket {
	rl.bucketMutex.RLock()
	bucket, exists := rl.buckets[ip]
	rl.bucketMutex.RUnlock()

	if exists {
		return bucket
	}

	// 创建新的令牌桶
	newBucket := NewTokenBucket(rl.capacity, rl.refillRate)

	rl.bucketMutex.Lock()
	// 再次检查，防止竞态条件
	if bucket, exists := rl.buckets[ip]; exists {
		rl.bucketMutex.Unlock()
		return bucket
	}

	rl.buckets[ip] = newBucket
	rl.bucketMutex.Unlock()

	return newBucket
}

// DefaultRateLimiter 默认速率限制器实例
var DefaultRateLimiter = NewRateLimiter(100, 10) // 默认100个令牌，每秒补充10个

// RateLimit API速率限制中间件
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取客户端IP
		clientIP := c.ClientIP()

		// 获取或创建令牌桶
		bucket := DefaultRateLimiter.GetBucket(clientIP)

		// 尝试获取令牌
		if !bucket.GetToken() {
			// 记录速率限制日志（如果日志系统已初始化）
			defer func() {
				if r := recover(); r != nil {
					// 忽略日志初始化失败的情况
				}
			}()
			logger.Warn("API速率限制触发", zap.String("client_ip", clientIP), zap.String("path", c.Request.URL.Path))

			// 返回429 Too Many Requests
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "请求频率过高，请稍后再试",
			})
			c.Abort()
			return
		}

		// 继续处理请求
		c.Next()
	}
}
