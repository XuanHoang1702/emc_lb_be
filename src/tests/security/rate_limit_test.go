package security_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"emc_lb/src/internal/middleware"
	"emc_lb/src/pkg/res"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestAuthRateLimit_FailClosedWhenRedisDown(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a Redis client pointing to a dummy port to simulate Redis being down
	rdb := redis.NewClient(&redis.Options{
		Addr:        "127.0.0.1:9999", // Invalid port
		DialTimeout: 10 * time.Millisecond,
		ReadTimeout: 10 * time.Millisecond,
	})

	router := gin.New()
	router.Use(middleware.AuthRateLimitMiddleware(rdb, 5, time.Minute, middleware.PolicyFailClosed))
	router.GET("/auth", func(c *gin.Context) {
		res.Success(c, http.StatusOK, "success")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/auth", nil)
	router.ServeHTTP(w, req)

	// Since Redis is down and policy is FailClosed, we expect a 503 Service Unavailable
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestAuthRateLimit_FailOpenWhenRedisDown(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a Redis client pointing to a dummy port to simulate Redis being down
	rdb := redis.NewClient(&redis.Options{
		Addr:        "127.0.0.1:9999", // Invalid port
		DialTimeout: 10 * time.Millisecond,
		ReadTimeout: 10 * time.Millisecond,
	})

	router := gin.New()
	router.Use(middleware.AuthRateLimitMiddleware(rdb, 5, time.Minute, middleware.PolicyFailOpen))
	router.GET("/payment/ipn", func(c *gin.Context) {
		res.Success(c, http.StatusOK, "success")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/payment/ipn", nil)
	router.ServeHTTP(w, req)

	// Since Redis is down and policy is FailOpen, we expect a 200 OK
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimit_FailOpenWhenRedisDown(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rdb := redis.NewClient(&redis.Options{
		Addr:        "127.0.0.1:9999", // Invalid port
		DialTimeout: 10 * time.Millisecond,
		ReadTimeout: 10 * time.Millisecond,
	})

	router := gin.New()
	router.Use(middleware.RateLimitMiddleware(rdb, 5, time.Minute))
	router.GET("/public", func(c *gin.Context) {
		res.Success(c, http.StatusOK, "success")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/public", nil)
	router.ServeHTTP(w, req)

	// Public endpoints are FailOpen
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimitByUser_FailOpenWhenRedisDown(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rdb := redis.NewClient(&redis.Options{
		Addr:        "127.0.0.1:9999", // Invalid port
		DialTimeout: 10 * time.Millisecond,
		ReadTimeout: 10 * time.Millisecond,
	})

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, "user-123")
		c.Next()
	})
	router.Use(middleware.RateLimitByUserMiddleware(rdb, 5, time.Minute))
	router.GET("/protected", func(c *gin.Context) {
		res.Success(c, http.StatusOK, "success")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	router.ServeHTTP(w, req)

	// Protected endpoints are FailOpen
	assert.Equal(t, http.StatusOK, w.Code)
}
