// Package interaction 提供用户交互状态相关的 HTTP 处理层。
// 聚合查询用户对视频的点赞、投币、收藏等交互状态。
package interaction

import (
	"net/http"
	"strconv"

	"github.com/My-TuDo/B-B/backend/internal/middleware"
	interactionservice "github.com/My-TuDo/B-B/backend/internal/service/interaction"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler 交互处理器，持有交互服务的引用。
type Handler struct {
	svc *interactionservice.Service
}

// NewHandler 创建交互处理器实例。
func NewHandler(svc *interactionservice.Service) *Handler {
	return &Handler{svc: svc}
}

// GetVideoInteractions 获取当前用户对指定视频的所有交互状态。
// GET /api/v1/videos/:id/interactions
// 需要登录认证，返回点赞、投币、收藏等状态汇总。
func (h *Handler) GetVideoInteractions(c *gin.Context) {
	// 获取当前登录用户 ID
	userID := middleware.GetUserID(c)

	// 解析视频 ID
	videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "视频ID格式错误")
		return
	}

	// 查询聚合交互状态
	status, err := h.svc.GetVideoInteractions(c.Request.Context(), userID, uint(videoID))
	if err != nil {
		logger.Error("get video interactions failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, status)
}
