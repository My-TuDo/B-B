// Package quality 提供视频画质相关的 HTTP 路由注册。
package quality

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes 注册画质相关路由到指定的路由组。
// 初始化 repository → service → handler 依赖链。
func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	repo := NewRepository(db)
	svc := NewService(repo)
	handler := NewHandler(svc)

	// 视频画质查询路由
	videos := r.Group("/videos")
	{
		videos.GET("/:id/qualities", handler.GetQualities)
	}
}
