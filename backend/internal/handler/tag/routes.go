package tag

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	tagrepo "github.com/My-TuDo/B-B/backend/internal/repository/tag"
	videorepo "github.com/My-TuDo/B-B/backend/internal/repository/video"
	tagservice "github.com/My-TuDo/B-B/backend/internal/service/tag"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	repo := tagrepo.NewRepository(db)
	videoRepo := videorepo.NewRepository(db)
	svc := tagservice.NewService(repo, videoRepo)
	handler := NewHandler(svc)

	tags := r.Group("/tags")
	{
		tags.GET("/", handler.List)
		tags.POST("/", middleware.AuthRequired(), handler.Create)
	}

	videos := r.Group("/videos")
	{
		videos.POST("/:id/tags", middleware.AuthRequired(), handler.SetVideoTags)
		videos.GET("/:id/tags", handler.GetVideoTags)
	}
}
