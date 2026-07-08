package coin

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	coinrepo "github.com/My-TuDo/B-B/backend/internal/repository/coin"
	coinservice "github.com/My-TuDo/B-B/backend/internal/service/coin"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, rdb *redis.Client) {
	repo := coinrepo.NewRepository(db)
	svc := coinservice.NewService(repo, rdb)
	handler := NewHandler(svc)

	r.POST("/videos/:id/coin", middleware.AuthRequired(), handler.AddCoin)
}
