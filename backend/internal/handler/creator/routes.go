package creator

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	creatorservice "github.com/My-TuDo/B-B/backend/internal/service/creator"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, creatorSvc *creatorservice.Service) {
	handler := NewHandler(creatorSvc)

	creator := r.Group("/creator")
	{
		creator.GET("/videos", middleware.AuthRequired(), handler.CreatorVideos)
		creator.GET("/stats", middleware.AuthRequired(), handler.CreatorStats)
	}
}
