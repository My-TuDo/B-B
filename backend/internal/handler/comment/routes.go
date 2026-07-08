package comment

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	commentrepo "github.com/My-TuDo/B-B/backend/internal/repository/comment"
	commentservice "github.com/My-TuDo/B-B/backend/internal/service/comment"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, rdb *redis.Client, notifier commentservice.Notifier) {
	repo := commentrepo.NewRepository(db)
	svc := commentservice.NewService(repo, rdb, notifier)
	handler := NewHandler(svc)

	videos := r.Group("/videos")
	{
		videos.GET("/:id/comments", handler.GetComments)
		videos.POST("/:id/comments", middleware.AuthRequired(), handler.CreateComment)
		videos.DELETE("/:id/comments/:comment_id", middleware.AuthRequired(), handler.DeleteComment)
	}

	r.POST("/comments/:id/like", middleware.AuthRequired(), handler.LikeComment)
}
