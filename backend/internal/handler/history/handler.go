// Package history 提供观看历史相关的 HTTP 处理层。
// 处理用户观看记录的创建/更新和分页查询。
package history

import (
	"net/http"
	"strconv"

	"github.com/My-TuDo/B-B/backend/internal/middleware"
	historymodel "github.com/My-TuDo/B-B/backend/internal/model/history"
	historyservice "github.com/My-TuDo/B-B/backend/internal/service/history"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler 历史记录处理器，持有历史记录服务的引用。
type Handler struct {
	svc *historyservice.Service
}

// NewHandler 创建历史记录处理器实例。
func NewHandler(svc *historyservice.Service) *Handler {
	return &Handler{svc: svc}
}

// CreateOrUpdate 创建或更新观看历史记录。
// POST /api/v1/history
// 需要登录认证，同一视频重复观看会更新进度而非重复插入。
func (h *Handler) CreateOrUpdate(c *gin.Context) {
	// 获取当前登录用户 ID
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	// 绑定请求体
	var req historymodel.CreateHistoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请求参数格式错误")
		return
	}

	if err := h.svc.CreateOrUpdate(c.Request.Context(), userID, &req); err != nil {
		logger.Error("create/update history failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, nil)
}

// List 获取当前用户的观看历史列表（分页、按时间倒序）。
// GET /api/v1/history
// 需要登录认证。
func (h *Handler) List(c *gin.Context) {
	// 获取当前登录用户 ID
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))

	resp, err := h.svc.List(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		logger.Error("list history failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}
