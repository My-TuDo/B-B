// Package tag 提供标签相关的 HTTP 处理层。
// 处理标签的创建、列表查询以及视频标签的绑定与查询。
package tag

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/My-TuDo/B-B/backend/internal/middleware"
	tagmodel "github.com/My-TuDo/B-B/backend/internal/model/tag"
	tagservice "github.com/My-TuDo/B-B/backend/internal/service/tag"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler 标签处理器，持有标签服务的引用。
type Handler struct {
	svc *tagservice.Service
}

// NewHandler 创建标签处理器实例。
func NewHandler(svc *tagservice.Service) *Handler {
	return &Handler{svc: svc}
}

// Create 创建新的标签。
// POST /api/v1/tags
// 需要登录认证，标签名不能为空且不超过 30 字符。
func (h *Handler) Create(c *gin.Context) {
	// 绑定请求体
	var req tagmodel.CreateTagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请求参数格式错误")
		return
	}

	// 校验标签名：非空且不超过 30 字符
	name := req.Name
	if name == "" || len([]rune(name)) > 30 {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "标签名不能为空且不超过30字符")
		return
	}

	resp, err := h.svc.Create(c.Request.Context(), name)
	if err != nil {
		logger.Error("create tag failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Created(c, resp)
}

// List 获取所有已创建的标签列表。
// GET /api/v1/tags
func (h *Handler) List(c *gin.Context) {
	resp, err := h.svc.List(c.Request.Context())
	if err != nil {
		logger.Error("list tags failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}

// SetVideoTags 设置指定视频的标签（覆盖式更新）。
// POST /api/v1/videos/:id/tags
// 需要登录认证，只有视频所有者可以设置。
func (h *Handler) SetVideoTags(c *gin.Context) {
	// 获取当前登录用户 ID
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	// 解析视频 ID
	videoIDStr := c.Param("id")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "视频ID格式错误")
		return
	}

	// 绑定请求体
	var req tagmodel.SetVideoTagsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请求参数格式错误")
		return
	}

	if err := h.svc.SetVideoTags(c.Request.Context(), userID, uint(videoID), &req); err != nil {
		// 区分服务层错误：视频不存在返回 404，非所有者返回 403
		var svcErr *tagservice.Error
		if errors.As(err, &svcErr) {
			status := http.StatusNotFound
			if svcErr.Code == errcode.Forbidden {
				status = http.StatusForbidden
			}
			response.Error(c, status, svcErr.Code, svcErr.Msg)
			return
		}
		logger.Error("set video tags failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, nil)
}

// GetVideoTags 获取指定视频的所有标签。
// GET /api/v1/videos/:id/tags
func (h *Handler) GetVideoTags(c *gin.Context) {
	// 解析视频 ID
	videoIDStr := c.Param("id")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "视频ID格式错误")
		return
	}

	resp, err := h.svc.GetVideoTags(c.Request.Context(), uint(videoID))
	if err != nil {
		logger.Error("get video tags failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}
