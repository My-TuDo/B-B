package user

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	userrepo "github.com/My-TuDo/B-B/backend/internal/repository/user"
	userservice "github.com/My-TuDo/B-B/backend/internal/service/user"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, rdb *redis.Client) {
	repo := userrepo.NewRepository(db)
	svc := userservice.NewService(repo)
	handler := NewHandler(svc)

	users := r.Group("/users")
	{
		users.GET("/:id", handler.GetUser)
		users.PUT("/:id", middleware.AuthRequired(), handler.UpdateUser)
		users.POST("/:id/avatar", middleware.AuthRequired(), handler.UploadAvatar)
	}
}
