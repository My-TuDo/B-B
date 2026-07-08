package admin

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	videorepo "github.com/My-TuDo/B-B/backend/internal/repository/video"
	adminservice "github.com/My-TuDo/B-B/backend/internal/service/admin"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, adminSvc *adminservice.Service) {
	videoRepo := videorepo.NewRepository(db)
	handler := NewHandler(videoRepo, adminSvc)

	admin := r.Group("/admin")
	{
		admin.GET("/videos", middleware.AuthRequired(), handler.AdminVideos)
		admin.PUT("/videos/:id/review", middleware.AuthRequired(), handler.Review)
	}
}
