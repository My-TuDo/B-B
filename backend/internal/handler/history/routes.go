// Package history 提供观看历史相关的 HTTP 路由注册。
package history

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	historyrepo "github.com/My-TuDo/B-B/backend/internal/repository/history"
	videorepo "github.com/My-TuDo/B-B/backend/internal/repository/video"
	historyservice "github.com/My-TuDo/B-B/backend/internal/service/history"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes 注册观看历史相关路由到指定的路由组。
// 初始化 repository（历史+视频）→ service → handler 依赖链。
func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	repo := historyrepo.NewRepository(db)
	videoRepo := videorepo.NewRepository(db)
	svc := historyservice.NewService(repo, videoRepo)
	handler := NewHandler(svc)

	// 历史记录路由组：需要认证
	history := r.Group("/history")
	{
		history.GET("/", middleware.AuthRequired(), handler.List)
		history.POST("/", middleware.AuthRequired(), handler.CreateOrUpdate)
	}
}
