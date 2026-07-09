package quality

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	repo := NewRepository(db)
	svc := NewService(repo)
	handler := NewHandler(svc)

	videos := r.Group("/videos")
	{
		videos.GET("/:id/qualities", handler.GetQualities)
	}
}
