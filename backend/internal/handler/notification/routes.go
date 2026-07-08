package notification

import (
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	messageservice "github.com/My-TuDo/B-B/backend/internal/service/message"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, svc *messageservice.Service) {
	handler := NewHandler(svc)

	notifications := r.Group("/notifications")
	{
		notifications.GET("/", middleware.AuthRequired(), handler.GetNotifications)
		notifications.POST("/read-all", middleware.AuthRequired(), handler.ReadAll)
		notifications.POST("/:id/read", middleware.AuthRequired(), handler.MarkSingleRead)
	}
}
