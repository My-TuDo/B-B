// Package comment 提供评论相关的路由注册。
package comment

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	commentrepo "github.com/My-TuDo/B-B/backend/internal/repository/comment"
	commentservice "github.com/My-TuDo/B-B/backend/internal/service/comment"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// RegisterRoutes 注册评论模块的所有路由到指定路由组。
// notifier 为通知接口，用于评论创建后触发通知（如 WebSocket 推送）。
func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, rdb *redis.Client, notifier commentservice.Notifier) {
	// 构建依赖链：Repository → Service → Handler
	repo := commentrepo.NewRepository(db)
	svc := commentservice.NewService(repo, rdb, notifier)
	handler := NewHandler(svc)

	videos := r.Group("/videos")
	{
		// 公开接口
		videos.GET("/:id/comments", handler.GetComments) // 获取视频评论列表

		// 需认证接口
		videos.POST("/:id/comments", middleware.AuthRequired(), handler.CreateComment)              // 创建评论
		videos.DELETE("/:id/comments/:comment_id", middleware.AuthRequired(), handler.DeleteComment) // 删除评论
	}

	// 评论点赞（独立路由）
	r.POST("/comments/:id/like", middleware.AuthRequired(), handler.LikeComment)
}
