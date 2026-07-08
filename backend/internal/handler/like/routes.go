package like

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	likerepo "github.com/My-TuDo/B-B/backend/internal/repository/like"
	likeservice "github.com/My-TuDo/B-B/backend/internal/service/like"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, rdb *redis.Client, messageSvc likeservice.Notifier) {
	repo := likerepo.NewRepository(db)
	svc := likeservice.NewService(repo, rdb, messageSvc)
	handler := NewHandler(svc)

	r.POST("/videos/:id/like", middleware.AuthRequired(), handler.ToggleLike)
}
