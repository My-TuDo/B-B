package middleware

import (
	"time"

	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		requestID, _ := c.Get("requestId")
		rid := ""
		if requestID != nil {
			rid = requestID.(string)
		}

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
