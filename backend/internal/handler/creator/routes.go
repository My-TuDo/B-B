// Package creator 提供创作者中心相关的 HTTP 路由注册。
package creator

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	creatorservice "github.com/My-TuDo/B-B/backend/internal/service/creator"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册创作者中心相关路由到指定的路由组。
// 直接接收已初始化的 creator service，无需在此重新组装依赖。
func RegisterRoutes(r *gin.RouterGroup, creatorSvc *creatorservice.Service) {
	handler := NewHandler(creatorSvc)

	// 创作者路由组：需要认证
	creator := r.Group("/creator")
	{
		creator.GET("/videos", middleware.AuthRequired(), handler.CreatorVideos)
		creator.GET("/stats", middleware.AuthRequired(), handler.CreatorStats)
	}
}
