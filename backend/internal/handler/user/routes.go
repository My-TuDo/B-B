// Package user 提供用户信息相关的路由注册。
package user

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	userrepo "github.com/My-TuDo/B-B/backend/internal/repository/user"
	userservice "github.com/My-TuDo/B-B/backend/internal/service/user"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// RegisterRoutes 注册用户模块的所有路由到指定路由组。
// 包括查看用户信息、更新个人信息、上传头像等接口。
func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, rdb *redis.Client) {
	// 构建依赖链：Repository → Service → Handler
	repo := userrepo.NewRepository(db)
	svc := userservice.NewService(repo)
	handler := NewHandler(svc)

	users := r.Group("/users")
	{
		// 公开接口
		users.GET("/:id", handler.GetUser) // 获取用户公开信息

		// 需认证接口
		users.PUT("/:id", middleware.AuthRequired(), handler.UpdateUser)                // 更新个人信息
		users.POST("/:id/avatar", middleware.AuthRequired(), handler.UploadAvatar)      // 上传头像
	}
}
