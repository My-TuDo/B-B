package interaction

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	interactionrepo "github.com/My-TuDo/B-B/backend/internal/repository/interaction"
	interactionservice "github.com/My-TuDo/B-B/backend/internal/service/interaction"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, rdb *redis.Client) {
	repo := interactionrepo.NewRepository(db)
	svc := interactionservice.NewService(repo, rdb)
	handler := NewHandler(svc)

	r.GET("/videos/:id/interactions", middleware.AuthRequired(), handler.GetVideoInteractions)
}
