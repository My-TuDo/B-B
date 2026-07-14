// Package danmaku 提供弹幕相关的 HTTP 处理器，包括弹幕列表获取、
// 发送弹幕以及 WebSocket 实时弹幕连接管理。
package danmaku

import (
	"net/http"
	"strconv"

	"github.com/My-TuDo/B-B/backend/internal/middleware"
	danmakumodel "github.com/My-TuDo/B-B/backend/internal/model/danmaku"
	danmakuservice "github.com/My-TuDo/B-B/backend/internal/service/danmaku"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler 是弹幕模块的 HTTP 处理器，持有弹幕服务实例。
type Handler struct {
	svc *danmakuservice.Service
}

// NewHandler 创建弹幕处理器实例。
func NewHandler(svc *danmakuservice.Service) *Handler {
	return &Handler{svc: svc}
}

// GetDanmaku 获取指定视频的所有弹幕列表（公开接口）。
func (h *Handler) GetDanmaku(c *gin.Context) {
	videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "视频ID格式错误")
		return
	}

	resps, err := h.svc.GetDanmaku(c.Request.Context(), uint(videoID))
	if err != nil {
		logger.Error("get danmaku failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resps)
}

// SendDanmaku 发送一条弹幕（需认证）。
// 弹幕内容长度限制为 1-200 字符，包含文本、出现时间和颜色等信息。
func (h *Handler) SendDanmaku(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "视频ID格式错误")
		return
	}

	var req danmakumodel.DanmakuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请求参数格式错误")
		return
	}

	// 弹幕内容长度校验（1-200 字符）
	if req.Content == "" || len(req.Content) > 200 {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "弹幕内容1-200字符")
		return
	}

	resp, err := h.svc.SendDanmaku(c.Request.Context(), uint(videoID), userID, &req)
	if err != nil {
		logger.Error("send danmaku failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Created(c, resp)
}
