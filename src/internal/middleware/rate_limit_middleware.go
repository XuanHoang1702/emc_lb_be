package middleware

import (
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"

	"emc_lb/src/pkg/res"
)

type RateLimitPolicy string

const (
	PolicyFailOpen   RateLimitPolicy = "fail_open"
	PolicyFailClosed RateLimitPolicy = "fail_closed"
)

var (
	rateLimitChecksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_checks_total",
			Help: "Total number of rate limit checks, categorized by limiter type and result.",
		},
		[]string{"limiter", "result"}, // result: allowed, rejected, error_fail_open, error_fail_closed
	)
)

// RateLimitMiddleware uses Redis to limit requests per IP address.
// Best for public/unauthenticated endpoints.
func RateLimitMiddleware(redisClient *redis.Client, rate int, per time.Duration) gin.HandlerFunc {
	limiter := redis_rate.NewLimiter(redisClient)

	return func(ctx *gin.Context) {
		ip := ctx.ClientIP()
		key := "rate_limit:" + ip

		limit := redis_rate.Limit{
			Rate:   rate,
			Burst:  rate,
			Period: per,
		}

		resResult, err := limiter.Allow(context.Background(), key, limit)
		if err != nil {
			// If Redis is down, we might want to either block or allow.
			// Allowing fallback is safer for availability, but we log the error.
			rateLimitChecksTotal.WithLabelValues("public_ip", "error_fail_open").Inc()
			ctx.Next()
			return
		}

		if resResult.Allowed == 0 {
			rateLimitChecksTotal.WithLabelValues("public_ip", "rejected").Inc()
			setRateLimitHeaders(ctx, limit.Rate, resResult)
			abortWithTooManyRequests(ctx)
			return
		}

		rateLimitChecksTotal.WithLabelValues("public_ip", "allowed").Inc()
		setRateLimitHeaders(ctx, limit.Rate, resResult)
		ctx.Next()
	}
}

// RateLimitByUserMiddleware uses Redis to limit requests per authenticated user ID.
// Falls back to IP-based limiting if user ID is not present.
// Best for protected/authenticated endpoints.
func RateLimitByUserMiddleware(redisClient *redis.Client, rate int, per time.Duration) gin.HandlerFunc {
	limiter := redis_rate.NewLimiter(redisClient)

	return func(ctx *gin.Context) {
		// Try user ID first (set by AccessTokenMiddleware), fall back to IP
		key := "rate_limit:user:"
		if userID, exists := ctx.Get(ContextUserIDKey); exists {
			key += userID.(string)
		} else {
			key = "rate_limit:ip:" + ctx.ClientIP()
		}

		limit := redis_rate.Limit{
			Rate:   rate,
			Burst:  rate,
			Period: per,
		}

		resResult, err := limiter.Allow(context.Background(), key, limit)
		if err != nil {
			rateLimitChecksTotal.WithLabelValues("user", "error_fail_open").Inc()
			ctx.Next()
			return
		}

		if resResult.Allowed == 0 {
			rateLimitChecksTotal.WithLabelValues("user", "rejected").Inc()
			setRateLimitHeaders(ctx, limit.Rate, resResult)
			abortWithTooManyRequests(ctx)
			return
		}

		rateLimitChecksTotal.WithLabelValues("user", "allowed").Inc()
		setRateLimitHeaders(ctx, limit.Rate, resResult)
		ctx.Next()
	}
}

// AuthRateLimitMiddleware applies strict rate limiting for sensitive auth endpoints
// (login, register, OTP verification, etc.) using IP + endpoint path as key.
// Much lower limits to prevent brute-force and credential stuffing attacks.
func AuthRateLimitMiddleware(redisClient *redis.Client, rate int, per time.Duration, policy RateLimitPolicy) gin.HandlerFunc {
	limiter := redis_rate.NewLimiter(redisClient)

	return func(ctx *gin.Context) {
		ip := ctx.ClientIP()
		path := ctx.FullPath()
		if path == "" {
			path = "unknown"
		}
		key := "rate_limit:auth:" + ip + ":" + path

		limit := redis_rate.Limit{
			Rate:   rate,
			Burst:  rate,
			Period: per,
		}

		resResult, err := limiter.Allow(context.Background(), key, limit)
		if err != nil {
			if policy == PolicyFailClosed {
				rateLimitChecksTotal.WithLabelValues("auth", "error_fail_closed").Inc()
				abortWithServiceUnavailable(ctx)
				return
			}
			rateLimitChecksTotal.WithLabelValues("auth", "error_fail_open").Inc()
			ctx.Next()
			return
		}

		if resResult.Allowed == 0 {
			rateLimitChecksTotal.WithLabelValues("auth", "rejected").Inc()
			setRateLimitHeaders(ctx, limit.Rate, resResult)
			abortWithTooManyRequests(ctx)
			return
		}

		rateLimitChecksTotal.WithLabelValues("auth", "allowed").Inc()
		setRateLimitHeaders(ctx, limit.Rate, resResult)
		ctx.Next()
	}
}

// setRateLimitHeaders sets standard rate limit response headers.
func setRateLimitHeaders(ctx *gin.Context, rate int, result *redis_rate.Result) {
	ctx.Header("RateLimit-Limit", strconv.Itoa(rate))
	ctx.Header("RateLimit-Remaining", strconv.Itoa(result.Remaining))
	ctx.Header("RateLimit-Reset", strconv.FormatInt(time.Now().Add(result.ResetAfter).Unix(), 10))
}

// abortWithTooManyRequests aborts the request with a 429 Too Many Requests response.
func abortWithTooManyRequests(ctx *gin.Context) {
	errRes := res.NewError("err_too_many_requests", res.ErrCodeTooManyRequests)
	ctx.Abort()
	res.Error(ctx, errRes)
}

// abortWithServiceUnavailable aborts the request with a 503 Service Unavailable response.
func abortWithServiceUnavailable(ctx *gin.Context) {
	errRes := res.NewError("err_service_unavailable", res.ErrCodeServiceUnavailable)
	ctx.Abort()
	res.Error(ctx, errRes)
}
