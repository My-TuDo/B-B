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
func StreamProgress(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid video id")
		return
	}
	videoID := uint(id)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.String(http.StatusInternalServerError, "streaming not supported")
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)

	broker := worker.GetBroker()
	ch := broker.Subscribe(videoID)
	defer broker.Unsubscribe(videoID, ch)

	ctx := c.Request.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case update, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(update)
			fmt.Fprintf(c.Writer, "data: %s\n\n", string(data))
			flusher.Flush()

			if update.Status == 2 { // done
				return
			}
		}
	}
}
