// Package like 提供点赞相关的 HTTP 路由注册。
package like

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	likerepo "github.com/My-TuDo/B-B/backend/internal/repository/like"
	likeservice "github.com/My-TuDo/B-B/backend/internal/service/like"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// RegisterRoutes 注册点赞相关路由到指定的路由组。
// 初始化 repository → service（含 Redis 缓存和通知器）→ handler 依赖链。
func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, rdb *redis.Client, messageSvc likeservice.Notifier) {
	repo := likerepo.NewRepository(db)
	svc := likeservice.NewService(repo, rdb, messageSvc)
	handler := NewHandler(svc)

	// 点赞切换路由：需要认证
	r.POST("/videos/:id/like", middleware.AuthRequired(), handler.ToggleLike)
}
