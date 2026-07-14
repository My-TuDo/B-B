// Package tag 提供标签相关的 HTTP 路由注册。
package tag

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	tagrepo "github.com/My-TuDo/B-B/backend/internal/repository/tag"
	videorepo "github.com/My-TuDo/B-B/backend/internal/repository/video"
	tagservice "github.com/My-TuDo/B-B/backend/internal/service/tag"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes 注册标签相关路由到指定的路由组。
// 初始化 repository（标签+视频）→ service → handler 依赖链。
func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	repo := tagrepo.NewRepository(db)
	videoRepo := videorepo.NewRepository(db)
	svc := tagservice.NewService(repo, videoRepo)
	handler := NewHandler(svc)

	// 标签路由组
	tags := r.Group("/tags")
	{
		tags.GET("/", handler.List)
		tags.POST("/", middleware.AuthRequired(), handler.Create)
	}

	// 视频标签路由组
	videos := r.Group("/videos")
	{
		videos.POST("/:id/tags", middleware.AuthRequired(), handler.SetVideoTags)
		videos.GET("/:id/tags", handler.GetVideoTags)
	}
}
