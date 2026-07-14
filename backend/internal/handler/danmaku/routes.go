// Package danmaku 提供弹幕相关的路由注册，包括 HTTP 接口和 WebSocket 连接。
package danmaku

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	danmakurepo "github.com/My-TuDo/B-B/backend/internal/repository/danmaku"
	danmakuservice "github.com/My-TuDo/B-B/backend/internal/service/danmaku"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// RegisterRoutes 注册弹幕模块的所有路由到指定路由组。
func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, rdb *redis.Client) {
	// 构建依赖链：Repository → Service → Handler
	repo := danmakurepo.NewRepository(db)
	svc := danmakuservice.NewService(repo, rdb)
	handler := NewHandler(svc)

	videos := r.Group("/videos")
	{
		// 公开接口：获取弹幕列表
		videos.GET("/:id/danmaku", handler.GetDanmaku)

		// 需认证接口：发送弹幕
		videos.POST("/:id/danmaku", middleware.AuthRequired(), handler.SendDanmaku)
	}

	// WebSocket 连接 — 公开读取（弹幕广播对所有用户可见）
	r.GET("/ws/danmaku/:video_id", handler.WebSocket)
}
