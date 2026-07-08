package favorite

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	favoriterepo "github.com/My-TuDo/B-B/backend/internal/repository/favorite"
	favoriteservice "github.com/My-TuDo/B-B/backend/internal/service/favorite"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	repo := favoriterepo.NewRepository(db)
	svc := favoriteservice.NewService(repo)
	handler := NewHandler(svc)

	favorites := r.Group("/favorites")
	{
		favorites.POST("/", middleware.AuthRequired(), handler.CreateFavorite)
		favorites.GET("/", middleware.AuthRequired(), handler.GetFavorites)
		favorites.GET("/:id", handler.GetFavoriteDetail)
		favorites.POST("/:id/items", middleware.AuthRequired(), handler.ToggleFavoriteItem)
	}

	// Public endpoint: user's public favorites (for profile page)
	r.GET("/users/:id/favorites", handler.GetUserFavorites)
}
