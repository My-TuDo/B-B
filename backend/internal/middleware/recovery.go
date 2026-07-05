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

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()
				requestID, _ := c.Get("requestId")
				rid := ""
				if requestID != nil {
					rid = requestID.(string)
				}

				logger.Error("panic recovered",
					zap.Any("error", err),
					zap.String("request_id", rid),
					zap.String("stack", string(stack)),
				)

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
