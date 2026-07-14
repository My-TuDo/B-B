// Package follow 提供关注/粉丝相关的 HTTP 处理层。
// 处理关注切换、粉丝列表、关注列表、动态推送和个人主页等操作。
package follow

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/My-TuDo/B-B/backend/internal/middleware"
	followservice "github.com/My-TuDo/B-B/backend/internal/service/follow"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler 关注处理器，持有关注服务的引用。
type Handler struct {
	svc            *followservice.Service
	interactionSvc interface{} // placeholder, used from main
}

// NewHandler 创建关注处理器实例。
func NewHandler(svc *followservice.Service) *Handler {
	return &Handler{svc: svc}
}

// ToggleFollow 切换关注状态：已关注则取消，未关注则关注。
// POST /api/v1/users/:id/follow
// 需要登录认证，不支持自己关注自己。
func (h *Handler) ToggleFollow(c *gin.Context) {
	// 获取当前登录用户 ID（关注者）
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	// 解析目标用户 ID（被关注者）
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "用户ID格式错误")
		return
	}

	resp, err := h.svc.ToggleFollow(c.Request.Context(), userID, uint(targetID))
	if err != nil {
		// 服务层错误（如自己关注自己）
		var svcErr *followservice.Error
		if errors.As(err, &svcErr) {
			response.Error(c, http.StatusBadRequest, svcErr.Code, svcErr.Msg)
			return
		}
		logger.Error("toggle follow failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}

// GetFollowers 获取指定用户的粉丝列表（分页）。
// GET /api/v1/users/:id/followers
func (h *Handler) GetFollowers(c *gin.Context) {
	// 解析目标用户 ID
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "用户ID格式错误")
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	resp, err := h.svc.GetFollowers(c.Request.Context(), uint(targetID), page, pageSize)
	if err != nil {
		logger.Error("get followers failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}

// GetFollowing 获取指定用户关注的人的列表（分页）。
// GET /api/v1/users/:id/following
func (h *Handler) GetFollowing(c *gin.Context) {
	// 解析目标用户 ID
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "用户ID格式错误")
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	resp, err := h.svc.GetFollowing(c.Request.Context(), uint(targetID), page, pageSize)
	if err != nil {
		logger.Error("get following failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}

// GetFeed 获取当前登录用户的关注动态（分页）。
// GET /api/v1/feed
// 需要登录认证，返回关注用户的最近视频列表。
func (h *Handler) GetFeed(c *gin.Context) {
	// 获取当前登录用户 ID
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))

	resp, err := h.svc.GetFeed(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		logger.Error("get feed failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}

// GetProfile 获取指定用户的个人主页信息（含关注状态、统计数据等）。
// GET /api/v1/users/:id/profile
// 支持未登录访问，登录时额外返回当前用户是否已关注该目标用户。
func (h *Handler) GetProfile(c *gin.Context) {
	// 解析目标用户 ID
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "用户ID格式错误")
		return
	}

	// 获取当前登录用户 ID（可能为 0，表示未登录）
	viewerID := middleware.GetUserID(c)

	resp, err := h.svc.GetProfile(c.Request.Context(), uint(targetID), viewerID)
	if err != nil {
		var svcErr *followservice.Error
		if errors.As(err, &svcErr) {
			response.Error(c, http.StatusNotFound, svcErr.Code, svcErr.Msg)
			return
		}
		logger.Error("get profile failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}
