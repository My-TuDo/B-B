// Package middleware 提供 Gin HTTP 中间件集合。
// 包含：Recovery（Panic 恢复）、RequestID（请求追踪）、Logger（请求日志）、
// CORS（跨域）、RateLimit（限流）、CSRF（跨站请求伪造防护）、Auth（JWT 认证）。
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

// globalLimiter 全局限流器：每秒 100 个令牌，突发容量 100。
// 使用令牌桶算法（Token Bucket），保护服务不被瞬时流量打垮。
// 该限流器对所有请求生效，无论是否认证。
var globalLimiter = rate.NewLimiter(rate.Limit(100), 100)

// RateLimit 返回限流中间件。
// 包含两级限流策略：
//  1. 全局令牌桶限流：所有请求共享，100 req/s
//  2. 认证接口 IP 级别限流：针对 /api/v1/auth/register 和 /api/v1/auth/login，
//     同一 IP 每分钟最多 5 次（通过 Redis 计数实现）
//
// 超限时返回 HTTP 429 Too Many Requests。
func RateLimit(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 第一级：全局令牌桶限流
		if !globalLimiter.Allow() {
			response.Error(c, http.StatusTooManyRequests, errcode.TooManyRequests, errcode.Message(errcode.TooManyRequests))
			c.Abort()
			return
		}

		// 第二级：认证接口 IP 限流（防暴力破解）
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

// ipRateLimitCheck 基于 Redis 的 IP 级别限流检查。
// 键格式：ratelimit:ip:{path}:{ip}，TTL 1 分钟，限制 5 次/分钟。
//
// 降级策略：Redis 不可用时放行（返回 true），避免 Redis 故障导致所有请求被拦截。
// INCR 是原子操作，保证并发安全。
func ipRateLimitCheck(rdb *redis.Client, ip, path string) bool {
	ctx := context.Background()
	key := fmt.Sprintf("ratelimit:ip:%s:%s", path, ip)

	// INCR 原子递增计数
	count, err := rdb.Incr(ctx, key).Result()
	if err != nil {
		// Redis 不可用时放行，避免误伤正常请求
		return true
	}

	// 首次访问时设置过期时间（1 分钟滑动窗口）
	if count == 1 {
		rdb.Expire(ctx, key, time.Minute)
	}

	return count <= 5 // 每分钟最多 5 次
}
