// Package notification 提供通知消息相关的 HTTP 处理层。
// 处理用户的通知列表查询、全部已读和单条已读操作。
package notification

import (
	"net/http"
	"strconv"

	"github.com/My-TuDo/B-B/backend/internal/middleware"
	messageservice "github.com/My-TuDo/B-B/backend/internal/service/message"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler 通知处理器，持有消息服务的引用。
type Handler struct {
	svc *messageservice.Service
}

// NewHandler 创建通知处理器实例。
func NewHandler(svc *messageservice.Service) *Handler {
	return &Handler{svc: svc}
}

// GetNotifications 获取当前用户的通知列表（分页）。
// GET /api/v1/notifications
// 需要登录认证。
func (h *Handler) GetNotifications(c *gin.Context) {
	// 获取当前登录用户 ID
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	resp, err := h.svc.GetNotifications(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		logger.Error("get notifications failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}

// ReadAll 将当前用户的所有未读通知标记为已读。
// POST /api/v1/notifications/read-all
// 需要登录认证。
func (h *Handler) ReadAll(c *gin.Context) {
	// 获取当前登录用户 ID
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	if err := h.svc.ReadAll(c.Request.Context(), userID); err != nil {
		logger.Error("read all notifications failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, nil)
}

// MarkSingleRead 将指定通知标记为已读。
// POST /api/v1/notifications/:id/read
// 需要登录认证，只能标记自己的通知。
func (h *Handler) MarkSingleRead(c *gin.Context) {
	// 获取当前登录用户 ID
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	// 解析通知 ID
	idStr := c.Param("id")
	messageID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "invalid id")
		return
	}

	if err := h.svc.MarkSingleRead(c.Request.Context(), userID, uint(messageID)); err != nil {
		logger.Error("mark single read failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, nil)
}
