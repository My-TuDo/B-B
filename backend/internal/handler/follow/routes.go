package follow

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	followrepo "github.com/My-TuDo/B-B/backend/internal/repository/follow"
	followservice "github.com/My-TuDo/B-B/backend/internal/service/follow"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, messageSvc followservice.Notifier) {
	repo := followrepo.NewRepository(db)
	svc := followservice.NewServiceWithNotifier(repo, db, messageSvc)
	handler := NewHandler(svc)

	users := r.Group("/users")
	{
		users.POST("/:id/follow", middleware.AuthRequired(), handler.ToggleFollow)
		users.GET("/:id/followers", handler.GetFollowers)
		users.GET("/:id/following", handler.GetFollowing)
		users.GET("/:id/profile", handler.GetProfile)
	}

	r.GET("/feed", middleware.AuthRequired(), handler.GetFeed)
}
