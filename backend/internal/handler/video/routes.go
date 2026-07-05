package video

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	videorepo "github.com/My-TuDo/B-B/backend/internal/repository/video"
	videoservice "github.com/My-TuDo/B-B/backend/internal/service/video"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, rdb *redis.Client) {
	repo := videorepo.NewRepository(db)
	svc := videoservice.NewService(repo)
	handler := NewHandler(svc)

	videos := r.Group("/videos")
	{
		// Public
		videos.GET("/", handler.ListVideos)
		videos.GET("/:id", handler.GetVideo)
		videos.GET("/:id/play-url", handler.GetPlayURL)
		videos.GET("/users/:id/videos", handler.ListUserVideos)

		// Auth required
		videos.POST("/", middleware.AuthRequired(), handler.Upload)
		videos.PUT("/:id", middleware.AuthRequired(), handler.UpdateVideo)
		videos.DELETE("/:id", middleware.AuthRequired(), handler.DeleteVideo)
	}
}
