// Package video 提供视频相关的路由注册。
package video

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	videorepo "github.com/My-TuDo/B-B/backend/internal/repository/video"
	videoservice "github.com/My-TuDo/B-B/backend/internal/service/video"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// RegisterRoutes 注册视频模块的所有路由到指定路由组。
// publishFn 为转码发布回调函数，用于视频上传成功后通知转码服务。
func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, rdb *redis.Client, publishFn func(videoID uint) error) {
	// 构建依赖链：Repository → Service → Handler
	repo := videorepo.NewRepository(db)
	svc := videoservice.NewServiceWithRedis(repo, rdb)
	svc.SetDB(db)
	if publishFn != nil {
		svc.SetTranscodePublisher(publishFn)
	}
	handler := NewHandler(svc)

	videos := r.Group("/videos")
	{
		// 公开接口
		videos.GET("/", handler.ListVideos)                     // 视频列表（支持分类筛选）
		videos.GET("/hot", handler.HotVideos)                    // 热门视频
		videos.GET("/ranking", handler.Ranking)                  // 排行榜（日/周/总）
		videos.GET("/:id", handler.GetVideo)                     // 视频详情
		videos.GET("/:id/play-url", handler.GetPlayURL)          // 播放地址
		videos.GET("/users/:id/videos", handler.ListUserVideos)  // 用户作品列表

		// 需认证接口
		videos.POST("/", middleware.AuthRequired(), handler.Upload)          // 上传视频（SSE 进度）
		videos.PUT("/:id", middleware.AuthRequired(), handler.UpdateVideo)   // 更新视频信息
		videos.DELETE("/:id", middleware.AuthRequired(), handler.DeleteVideo) // 删除视频
	}
}
