package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GT-610/tairitsu/internal/app/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenBucket_GetToken(t *testing.T) {
	// Arrange
	bucket := middleware.NewTokenBucket(5, 1) // 容量5，每秒补充1个

	// Act & Assert
	// 初始状态应该有5个令牌
	for i := 0; i < 5; i++ {
		assert.True(t, bucket.GetToken())
	}

	// 第6个请求应该失败
	assert.False(t, bucket.GetToken())

	// 等待1秒，应该补充1个令牌
	time.Sleep(time.Second)
	assert.True(t, bucket.GetToken())

	// 再次请求应该失败
	assert.False(t, bucket.GetToken())
}

func TestRateLimiter_GetBucket(t *testing.T) {
	// Arrange
	limiter := middleware.NewRateLimiter(10, 2)

	// Act
	bucket1 := limiter.GetBucket("192.168.1.1")
	bucket2 := limiter.GetBucket("192.168.1.2")
	bucket3 := limiter.GetBucket("192.168.1.1") // 同一个IP应该返回同一个桶

	// Assert
	assert.NotNil(t, bucket1)
	assert.NotNil(t, bucket2)
	assert.NotNil(t, bucket3)
	assert.NotEqual(t, bucket1, bucket2) // 不同IP应该有不同的桶
	assert.Equal(t, bucket1, bucket3)    // 同一个IP应该返回同一个桶
}

func TestRateLimitMiddleware(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.RateLimitWithLimiter(middleware.NewRateLimiter(1, 0)))
	app.Get("/limited", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	firstResp, err := app.Test(httptest.NewRequest(http.MethodGet, "/limited", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, firstResp.StatusCode)

	secondResp, err := app.Test(httptest.NewRequest(http.MethodGet, "/limited", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusTooManyRequests, secondResp.StatusCode)

	var body map[string]any
	require.NoError(t, json.NewDecoder(secondResp.Body).Decode(&body))
	assert.Equal(t, "system.rate_limited", body["error_code"])
}

func TestRateLimitMiddleware_WithWait(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.RateLimitWithLimiter(middleware.NewRateLimiter(1, 1)))
	app.Get("/limited", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	firstResp, err := app.Test(httptest.NewRequest(http.MethodGet, "/limited", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, firstResp.StatusCode)

	secondResp, err := app.Test(httptest.NewRequest(http.MethodGet, "/limited", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusTooManyRequests, secondResp.StatusCode)

	time.Sleep(1100 * time.Millisecond)

	thirdResp, err := app.Test(httptest.NewRequest(http.MethodGet, "/limited", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, thirdResp.StatusCode)
}
