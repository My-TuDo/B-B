// Package follow 提供关注/粉丝相关的 HTTP 路由注册。
package follow

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	followrepo "github.com/My-TuDo/B-B/backend/internal/repository/follow"
	followservice "github.com/My-TuDo/B-B/backend/internal/service/follow"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes 注册关注/粉丝相关路由到指定的路由组。
// 初始化 repository → service（含通知器）→ handler 依赖链。
func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, messageSvc followservice.Notifier) {
	repo := followrepo.NewRepository(db)
	svc := followservice.NewServiceWithNotifier(repo, db, messageSvc)
	handler := NewHandler(svc)

	// 用户相关路由组
	users := r.Group("/users")
	{
		users.POST("/:id/follow", middleware.AuthRequired(), handler.ToggleFollow)
		users.GET("/:id/followers", handler.GetFollowers)
		users.GET("/:id/following", handler.GetFollowing)
		users.GET("/:id/profile", handler.GetProfile)
	}

	// 关注动态路由：需要认证
	r.GET("/feed", middleware.AuthRequired(), handler.GetFeed)
}
