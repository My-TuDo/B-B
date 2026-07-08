package danmaku

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	danmakurepo "github.com/My-TuDo/B-B/backend/internal/repository/danmaku"
	danmakuservice "github.com/My-TuDo/B-B/backend/internal/service/danmaku"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, rdb *redis.Client) {
	repo := danmakurepo.NewRepository(db)
	svc := danmakuservice.NewService(repo, rdb)
	handler := NewHandler(svc)

	videos := r.Group("/videos")
	{
		videos.GET("/:id/danmaku", handler.GetDanmaku)
		videos.POST("/:id/danmaku", middleware.AuthRequired(), handler.SendDanmaku)
	}

	// WebSocket — public read (danmaku broadcast visible to all)
	r.GET("/ws/danmaku/:video_id", handler.WebSocket)
}
