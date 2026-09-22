package middleware

import (
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"

	"emc_lb/src/pkg/res"
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
			ctx.Next()
			return
		}

		setRateLimitHeaders(ctx, limit.Rate, resResult)

		if resResult.Allowed == 0 {
			abortWithTooManyRequests(ctx)
			return
		}

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
			ctx.Next()
			return
		}

		setRateLimitHeaders(ctx, limit.Rate, resResult)

		if resResult.Allowed == 0 {
			abortWithTooManyRequests(ctx)
			return
		}

		ctx.Next()
	}
}

// AuthRateLimitMiddleware applies strict rate limiting for sensitive auth endpoints
// (login, register, OTP verification, etc.) using IP + endpoint path as key.
// Much lower limits to prevent brute-force and credential stuffing attacks.
func AuthRateLimitMiddleware(redisClient *redis.Client, rate int, per time.Duration) gin.HandlerFunc {
	limiter := redis_rate.NewLimiter(redisClient)

	return func(ctx *gin.Context) {
		ip := ctx.ClientIP()
		path := ctx.FullPath()
		key := "rate_limit:auth:" + ip + ":" + path

		limit := redis_rate.Limit{
			Rate:   rate,
			Burst:  rate,
			Period: per,
		}

		resResult, err := limiter.Allow(context.Background(), key, limit)
		if err != nil {
			ctx.Next()
			return
		}

		setRateLimitHeaders(ctx, limit.Rate, resResult)

		if resResult.Allowed == 0 {
			abortWithTooManyRequests(ctx)
			return
		}

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
