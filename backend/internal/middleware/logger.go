package middleware

import (
	"time"

	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger 记录每个 HTTP 请求的方法、路径、状态码、耗时、客户端 IP 和请求 ID。
// 在请求处理完成后（c.Next() 返回后）记录。
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now() // 记录请求开始时间

		c.Next() // 执行后续 Handler

		// 计算耗时
		latency := time.Since(start)
		status := c.Writer.Status()

		// 获取请求 ID
		requestID, _ := c.Get("requestId")
		rid := ""
		if requestID != nil {
			rid = requestID.(string)
		}

		// 记录结构化日志
		logger.Info("request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", rid),
		)
	}
}
