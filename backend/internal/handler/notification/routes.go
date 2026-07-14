// Package notification 提供通知消息相关的 HTTP 路由注册。
package notification

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	messageservice "github.com/My-TuDo/B-B/backend/internal/service/message"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册通知相关路由到指定的路由组。
// 直接接收已初始化的 message service，无需在此重新组装依赖。
func RegisterRoutes(r *gin.RouterGroup, svc *messageservice.Service) {
	handler := NewHandler(svc)

	// 通知路由组：需要认证
	notifications := r.Group("/notifications")
	{
		notifications.GET("/", middleware.AuthRequired(), handler.GetNotifications)
		notifications.POST("/read-all", middleware.AuthRequired(), handler.ReadAll)
		notifications.POST("/:id/read", middleware.AuthRequired(), handler.MarkSingleRead)
	}
}
