package search

import (
	searchrepo "github.com/My-TuDo/B-B/backend/internal/repository/search"
	searchservice "github.com/My-TuDo/B-B/backend/internal/service/search"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	repo := searchrepo.NewRepository(db)
	svc := searchservice.NewService(repo)
	handler := NewHandler(svc)

	search := r.Group("/search")
	{
		search.GET("/", handler.Search)
		search.GET("/suggestions", handler.Suggestions)
	}
}
