package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

var globalLimiter = rate.NewLimiter(rate.Limit(100), 100)

func RateLimit(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Global token bucket
		if !globalLimiter.Allow() {
			response.Error(c, http.StatusTooManyRequests, errcode.TooManyRequests, errcode.Message(errcode.TooManyRequests))
			c.Abort()
			return
		}

		// IP rate limit for auth endpoints
		path := c.Request.URL.Path
		if path == "/api/v1/auth/register" || path == "/api/v1/auth/login" {
			if !ipRateLimitCheck(rdb, c.ClientIP(), path) {
				response.Error(c, http.StatusTooManyRequests, errcode.TooManyRequests, errcode.Message(errcode.TooManyRequests))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

func ipRateLimitCheck(rdb *redis.Client, ip, path string) bool {
	ctx := context.Background()
	key := fmt.Sprintf("ratelimit:ip:%s:%s", path, ip)

	count, err := rdb.Incr(ctx, key).Result()
	if err != nil {
		// Redis unavailable, allow through
		return true
	}

	if count == 1 {
		rdb.Expire(ctx, key, time.Minute)
	}

	return count <= 5
}
