package history

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	historyrepo "github.com/My-TuDo/B-B/backend/internal/repository/history"
	videorepo "github.com/My-TuDo/B-B/backend/internal/repository/video"
	historyservice "github.com/My-TuDo/B-B/backend/internal/service/history"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	repo := historyrepo.NewRepository(db)
	videoRepo := videorepo.NewRepository(db)
	svc := historyservice.NewService(repo, videoRepo)
	handler := NewHandler(svc)

	history := r.Group("/history")
	{
		history.GET("/", middleware.AuthRequired(), handler.List)
		history.POST("/", middleware.AuthRequired(), handler.CreateOrUpdate)
	}
}
