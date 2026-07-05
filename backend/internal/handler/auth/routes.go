package auth

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	authrepo "github.com/My-TuDo/B-B/backend/internal/repository/auth"
	authservice "github.com/My-TuDo/B-B/backend/internal/service/auth"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, rdb *redis.Client) {
	repo := authrepo.NewRepository(db)
	svc := authservice.NewService(repo, rdb)
	handler := NewHandler(svc)

	auth := r.Group("/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
		auth.POST("/logout", middleware.AuthRequired(), handler.Logout)
		auth.POST("/refresh", handler.Refresh)
		auth.GET("/me", middleware.AuthRequired(), handler.Me)
	}
}
