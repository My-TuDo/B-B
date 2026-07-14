// Package middleware 提供 Gin HTTP 中间件集合。
// 包含：Recovery（Panic 恢复）、RequestID（请求追踪）、Logger（请求日志）、
// CORS（跨域）、RateLimit（限流）、CSRF（跨站请求伪造防护）、Auth（JWT 认证）。
package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery 捕获 Handler 中的 panic，记录堆栈并返回 500 错误。
// 防止单个请求的 panic 导致整个服务崩溃。
//
// 行为：
//  1. 捕获 panic 并记录完整调用栈
//  2. 将错误详情写入结构化日志（含 request_id 关联）
//  3. 如果响应尚未写入，返回 HTTP 500 + 统一错误格式
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取完整调用栈（用于定位问题）
				stack := debug.Stack()

				// 尝试获取请求 ID（可能尚未注入，需要容错）
				requestID, _ := c.Get("requestId")
				rid := ""
				if requestID != nil {
					rid = requestID.(string)
				}

				// 记录错误日志（级别 Error，包含堆栈信息）
				logger.Error("panic recovered",
					zap.Any("error", err),
					zap.String("request_id", rid),
					zap.String("stack", string(stack)),
				)

				// 如果尚未写入响应，返回 500
				// 使用 ErrorWithRequestID 避免依赖 Context 中的 requestId
				if !c.Writer.Written() {
					response.ErrorWithRequestID(c, http.StatusInternalServerError,
						errcode.Internal,
						fmt.Sprintf("服务器内部错误 (RequestID: %s)", rid), rid)
				}
			}
		}()

		c.Next()
	}
}
