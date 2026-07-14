// Package transcode 提供视频转码进度 SSE 实时推送功能。
package transcode

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/My-TuDo/B-B/backend/internal/worker"
	"github.com/gin-gonic/gin"
)

// StreamProgress opens an SSE connection and streams real-time transcode progress.
// GET /api/v1/videos/:id/transcode-stream
// 建立 SSE 长连接，通过 worker broker 订阅转码进度更新并实时推送给前端。
// 当转码完成（status == 2）或客户端断开时自动关闭连接。
func StreamProgress(c *gin.Context) {
	// 解析视频 ID
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid video id")
		return
	}
	videoID := uint(id)

	// 确认客户端支持 SSE（需要 http.Flusher）
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.String(http.StatusInternalServerError, "streaming not supported")
		return
	}

	// 设置 SSE 响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no") // 禁用 nginx 缓冲
	c.Writer.WriteHeader(http.StatusOK)

	// 订阅该视频的转码进度更新
	broker := worker.GetBroker()
	ch := broker.Subscribe(videoID)
	defer broker.Unsubscribe(videoID, ch)

	ctx := c.Request.Context()
	for {
		select {
		case <-ctx.Done():
			// 客户端断开连接
			return
		case update, ok := <-ch:
			if !ok {
				// 通道关闭
				return
			}
			// 将进度更新序列化为 SSE 格式推送
			data, _ := json.Marshal(update)
			fmt.Fprintf(c.Writer, "data: %s\n\n", string(data))
			flusher.Flush()

			if update.Status == 2 { // done —— 转码完成，关闭连接
				return
			}
		}
	}
}
