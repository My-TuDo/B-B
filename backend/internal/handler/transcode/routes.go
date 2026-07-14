// Package transcode 提供视频转码状态相关的 HTTP 路由注册。
package transcode

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes 注册转码相关路由到指定的路由组。
// 初始化 repository → service → handler 依赖链。
func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	repo := NewRepository(db)
	svc := NewService(repo)
	handler := NewHandler(svc)

	// 视频转码路由组
	videos := r.Group("/videos")
	{
		videos.GET("/:id/transcode-status", handler.GetStatus)
		videos.GET("/:id/transcode-stream", StreamProgress)
	}
}
