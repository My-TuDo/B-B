// Package quality 提供视频画质相关的 HTTP 处理层。
// 视频转码后会生成多个清晰度版本，本包对外暴露各清晰度的播放地址。
package quality

import (
	"net/http"
	"strconv"

	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler 画质处理器，持有画质服务的引用。
type Handler struct {
	svc *Service
}

// NewHandler 创建画质处理器实例。
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GetQualities 获取指定视频的所有可用画质及对应的播放地址。
// GET /api/v1/videos/:id/qualities
func (h *Handler) GetQualities(c *gin.Context) {
	// 解析视频 ID
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "视频ID格式错误")
		return
	}

	// 查询该视频的所有画质版本
	qualities, err := h.svc.GetQualities(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("get qualities failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, qualities)
}
