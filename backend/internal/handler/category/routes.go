package category

import (
	categoryrepo "github.com/My-TuDo/B-B/backend/internal/repository/category"
	categoryservice "github.com/My-TuDo/B-B/backend/internal/service/category"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	repo := categoryrepo.NewRepository(db)
	svc := categoryservice.NewService(repo)
	handler := NewHandler(svc)

	categories := r.Group("/categories")
	{
		categories.GET("/", handler.List)
	}
}
