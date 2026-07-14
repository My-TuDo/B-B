// Package creator 提供创作者中心相关的 HTTP 处理层。
// 处理创作者查看自己的视频列表和统计数据。
package creator

import (
	"net/http"
	"strconv"

	"github.com/My-TuDo/B-B/backend/internal/middleware"
	creatorservice "github.com/My-TuDo/B-B/backend/internal/service/creator"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler 创作者处理器，持有创作者服务的引用。
type Handler struct {
	svc *creatorservice.Service
}

// NewHandler 创建创作者处理器实例。
func NewHandler(svc *creatorservice.Service) *Handler {
	return &Handler{svc: svc}
}

// CreatorVideos 获取创作者的视频列表（分页）。
// GET /api/v1/creator/videos
// 需要登录且角色为创作者（role >= 2）。
func (h *Handler) CreatorVideos(c *gin.Context) {
	// 获取当前登录用户 ID
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	// 权限校验：只有创作者及以上角色才能访问
	role := middleware.GetRole(c)
	if role < 2 {
		response.Error(c, http.StatusForbidden, errcode.Forbidden, errcode.Message(errcode.Forbidden))
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))

	resp, err := h.svc.ListVideos(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		logger.Error("creator videos failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}

// CreatorStats 获取创作者的统计数据（播放量、粉丝数等）。
// GET /api/v1/creator/stats
// 需要登录且角色为创作者（role >= 2）。
func (h *Handler) CreatorStats(c *gin.Context) {
	// 获取当前登录用户 ID
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	// 权限校验：只有创作者及以上角色才能访问
	role := middleware.GetRole(c)
	if role < 2 {
		response.Error(c, http.StatusForbidden, errcode.Forbidden, errcode.Message(errcode.Forbidden))
		return
	}

	resp, err := h.svc.GetStats(c.Request.Context(), userID)
	if err != nil {
		logger.Error("creator stats failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}
