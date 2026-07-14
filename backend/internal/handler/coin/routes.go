// Package coin 提供投币相关的 HTTP 路由注册。
package coin

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	coinrepo "github.com/My-TuDo/B-B/backend/internal/repository/coin"
	coinservice "github.com/My-TuDo/B-B/backend/internal/service/coin"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// RegisterRoutes 注册投币相关路由到指定的路由组。
// 初始化 repository → service → handler 依赖链。
// 投币操作需要登录认证和 Redis 用于限流/去重。
func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, rdb *redis.Client) {
	repo := coinrepo.NewRepository(db)
	svc := coinservice.NewService(repo, rdb)
	handler := NewHandler(svc)

	// 投币路由：需要认证
	r.POST("/videos/:id/coin", middleware.AuthRequired(), handler.AddCoin)
}
