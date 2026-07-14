package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID 为每个请求注入唯一追踪 ID。
// 优先使用客户端传入的 X-Request-Id 头，若无则生成 UUID v4。
// 该 ID 会写入上下文（c.Set）和响应头（X-Request-Id）。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优先从请求头获取
		requestID := c.GetHeader("X-Request-Id")
		if requestID == "" {
			requestID = uuid.New().String() // 生成新的 UUID
		}
		// 注入上下文，后续 Handler/中间件可通过 c.Get("requestId") 获取
		c.Set("requestId", requestID)
		// 写入响应头
		c.Header("X-Request-Id", requestID)
		c.Next()
	}
}
