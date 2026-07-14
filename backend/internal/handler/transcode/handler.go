// Package transcode 提供视频转码状态相关的 HTTP 处理层。
// 支持轮询转码状态和 SSE 实时推送转码进度。
package transcode

import (
	"net/http"
	"strconv"

	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler 转码处理器，持有转码服务的引用。
type Handler struct {
	svc *Service
}

// NewHandler 创建转码处理器实例。
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GetStatus returns the transcode status for a video.
// GET /api/v1/videos/:id/transcode-status
// 返回转码状态和进度百分比；若无记录则返回已完成状态。
func (h *Handler) GetStatus(c *gin.Context) {
	// 解析视频 ID
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "视频ID格式错误")
		return
	}

	// 查询转码任务状态
	task, err := h.svc.GetStatus(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("get transcode status failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, gin.H{
		"status":   task.Status,
		"progress": task.Progress,
	})
}
