// Package favorite 提供收藏夹相关的 HTTP 路由注册。
package favorite

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	favoriterepo "github.com/My-TuDo/B-B/backend/internal/repository/favorite"
	favoriteservice "github.com/My-TuDo/B-B/backend/internal/service/favorite"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes 注册收藏夹相关路由到指定的路由组。
// 初始化 repository → service → handler 依赖链。
func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	repo := favoriterepo.NewRepository(db)
	svc := favoriteservice.NewService(repo)
	handler := NewHandler(svc)

	// 收藏夹路由组
	favorites := r.Group("/favorites")
	{
		favorites.POST("/", middleware.AuthRequired(), handler.CreateFavorite)
		favorites.GET("/", middleware.AuthRequired(), handler.GetFavorites)
		favorites.GET("/:id", handler.GetFavoriteDetail)
		favorites.POST("/:id/items", middleware.AuthRequired(), handler.ToggleFavoriteItem)
	}

	// 公开端点：查看用户的公开收藏夹（用于个人主页）
	r.GET("/users/:id/favorites", handler.GetUserFavorites)
}
