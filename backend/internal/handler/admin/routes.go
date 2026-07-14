// Package admin 提供管理后台相关的路由注册。
package admin

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	videorepo "github.com/My-TuDo/B-B/backend/internal/repository/video"
	adminservice "github.com/My-TuDo/B-B/backend/internal/service/admin"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes 注册管理后台模块的所有路由到指定路由组。
// adminSvc 由上层传入，用于视频列表等管理服务。
func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, adminSvc *adminservice.Service) {
	// 构建依赖链：Repository → Handler
	videoRepo := videorepo.NewRepository(db)
	handler := NewHandler(db, videoRepo, adminSvc)

	admin := r.Group("/admin")
	{
		// 所有管理后台接口均需认证，内部再校验 role >= 3
		admin.GET("/stats", middleware.AuthRequired(), handler.Stats)           // 数据统计看板
		admin.GET("/users", middleware.AuthRequired(), handler.Users)           // 用户列表（支持搜索）
		admin.PUT("/users/:id/role", middleware.AuthRequired(), handler.UpdateUserRole) // 修改用户角色
		admin.GET("/videos", middleware.AuthRequired(), handler.AdminVideos)    // 视频管理列表
		admin.PUT("/videos/:id/review", middleware.AuthRequired(), handler.Review)       // 视频审核
		admin.GET("/system", middleware.AuthRequired(), handler.System)         // 系统信息
	}
}
