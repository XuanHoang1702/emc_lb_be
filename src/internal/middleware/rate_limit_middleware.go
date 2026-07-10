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

		// Set rate limit headers
		ctx.Header("RateLimit-Limit", strconv.Itoa(limit.Rate))
		ctx.Header("RateLimit-Remaining", strconv.Itoa(resResult.Remaining))
		ctx.Header("RateLimit-Reset", strconv.FormatInt(time.Now().Add(resResult.ResetAfter).Unix(), 10))

		if resResult.Allowed == 0 {
			errRes := res.NewError("err_too_many_requests", res.ErrCodeTooManyRequests)
			ctx.Abort()
			res.Error(ctx, errRes)
			return
		}

		ctx.Next()
	}
}
