package danmaku

import (
	"net/http"
	"strconv"

	"github.com/My-TuDo/B-B/backend/internal/middleware"
	"github.com/My-TuDo/B-B/backend/internal/ws"
	"github.com/gin-gonic/gin"
	gorillaWs "github.com/gorilla/websocket"
)

var upgrader = gorillaWs.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin in dev
	},
}

func (h *Handler) WebSocket(c *gin.Context) {
	videoIDStr := c.Param("video_id")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid video_id"})
		return
	}

	userID := middleware.GetUserID(c)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &ws.Client{
		Conn:    conn,
		Send:    make(chan []byte, 256),
		VideoID: uint(videoID),
		UserID:  userID,
		Hub:     ws.DefaultHub,
	}

	client.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
