// Package danmaku 提供弹幕 WebSocket 连接管理。
// 通过 gorilla/websocket 将 HTTP 连接升级为 WebSocket，
// 然后注册到全局 Hub 进行弹幕的实时广播与接收。
package danmaku

import (
	"net/http"
	"strconv"

	"github.com/My-TuDo/B-B/backend/internal/middleware"
	"github.com/My-TuDo/B-B/backend/internal/ws"
	"github.com/gin-gonic/gin"
	gorillaWs "github.com/gorilla/websocket"
)

// upgrader 是 WebSocket 升级器，配置读写缓冲区大小和跨域策略。
var upgrader = gorillaWs.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// CheckOrigin 允许所有来源的连接请求（开发环境，生产需收紧）
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSocket 处理弹幕 WebSocket 连接（公开接口）。
// 将 HTTP 连接升级为 WebSocket，创建客户端并注册到全局 Hub，
// 启动独立的读写协程进行双向实时通信。
func (h *Handler) WebSocket(c *gin.Context) {
	// 解析路由参数中的视频 ID
	videoIDStr := c.Param("video_id")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid video_id"})
		return
	}

	// 获取当前用户 ID（未登录则为 0）
	userID := middleware.GetUserID(c)

	// 将 HTTP 连接升级为 WebSocket 连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// 创建 WebSocket 客户端并注册到全局 Hub
	client := &ws.Client{
		Conn:    conn,
		Send:    make(chan []byte, 256), // 发送缓冲区
		VideoID: uint(videoID),
		UserID:  userID,
		Hub:     ws.DefaultHub,
	}

	// 将客户端注册到 Hub（Hub 按 videoID 管理房间）
	client.Hub.Register <- client

	// 启动独立的读写协程
	go client.WritePump() // 向 WebSocket 写数据（发送弹幕）
	go client.ReadPump()  // 从 WebSocket 读数据（接收弹幕）
}
