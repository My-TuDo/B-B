// Package like 提供点赞相关的 HTTP 处理层。
// 处理用户对视频的点赞/取消点赞切换操作。
package like

import (
	"net/http"
	"strconv"

	"github.com/My-TuDo/B-B/backend/internal/middleware"
	likeservice "github.com/My-TuDo/B-B/backend/internal/service/like"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler 点赞处理器，持有点赞服务的引用。
type Handler struct {
	svc *likeservice.Service
}

// NewHandler 创建点赞处理器实例。
func NewHandler(svc *likeservice.Service) *Handler {
	return &Handler{svc: svc}
}

// ToggleLike 切换点赞状态：已点赞则取消，未点赞则点赞。
// POST /api/v1/videos/:id/like
// 需要登录认证。
func (h *Handler) ToggleLike(c *gin.Context) {
	// 获取当前登录用户 ID
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	// 解析视频 ID
	videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "视频ID格式错误")
		return
	}

	// 执行点赞/取消切换
	resp, err := h.svc.ToggleLike(c.Request.Context(), userID, uint(videoID))
	if err != nil {
		logger.Error("toggle like failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}
