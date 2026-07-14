// Package auth 提供认证相关的路由注册。
package auth

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	authrepo "github.com/My-TuDo/B-B/backend/internal/repository/auth"
	authservice "github.com/My-TuDo/B-B/backend/internal/service/auth"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// RegisterRoutes 注册认证模块的所有路由到指定路由组。
// 包括注册、登录、登出、刷新 Token、个人信息等接口，以及 CSRF 令牌获取。
func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, rdb *redis.Client) {
	// 构建依赖链：Repository → Service → Handler
	repo := authrepo.NewRepository(db)
	svc := authservice.NewService(repo, rdb, db)
	handler := NewHandler(svc)

	auth := r.Group("/auth")
	{
		// 公开接口
		auth.POST("/register", handler.Register) // 用户注册
		auth.POST("/login", handler.Login)       // 用户登录
		auth.POST("/refresh", handler.Refresh)   // 刷新 Token

		// 需认证接口
		auth.POST("/logout", middleware.AuthRequired(), handler.Logout) // 用户登出
		auth.GET("/me", middleware.AuthRequired(), handler.Me)          // 获取当前用户信息
	}

	// CSRF 令牌获取（独立路由，不在 /auth 组下）
	r.GET("/csrf-token", handler.CSRF)
}
