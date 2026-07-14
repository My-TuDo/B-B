// Package favorite 提供收藏夹相关的 HTTP 处理层。
// 处理收藏夹的创建、查询、详情获取以及收藏项切换操作。
package favorite

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/My-TuDo/B-B/backend/internal/middleware"
	favoritemodel "github.com/My-TuDo/B-B/backend/internal/model/favorite"
	favoriteservice "github.com/My-TuDo/B-B/backend/internal/service/favorite"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler 收藏夹处理器，持有收藏夹服务的引用。
type Handler struct {
	svc *favoriteservice.Service
}

// NewHandler 创建收藏夹处理器实例。
func NewHandler(svc *favoriteservice.Service) *Handler {
	return &Handler{svc: svc}
}

// CreateFavorite 创建新的收藏夹。
// POST /api/v1/favorites
// 需要登录认证，收藏夹名称不能为空。
func (h *Handler) CreateFavorite(c *gin.Context) {
	// 获取当前登录用户 ID
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	// 绑定请求体
	var req favoritemodel.FavoriteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请求参数格式错误")
		return
	}

	// 校验收藏夹名称
	if req.Name == "" {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "收藏夹名称不能为空")
		return
	}

	resp, err := h.svc.CreateFavorite(c.Request.Context(), userID, &req)
	if err != nil {
		logger.Error("create favorite failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Created(c, resp)
}

// GetFavorites 获取当前用户的所有收藏夹列表。
// GET /api/v1/favorites
// 需要登录认证。
func (h *Handler) GetFavorites(c *gin.Context) {
	// 获取当前登录用户 ID
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	resps, err := h.svc.GetFavorites(c.Request.Context(), userID)
	if err != nil {
		logger.Error("get favorites failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resps)
}

// GetFavoriteDetail 获取指定收藏夹的详情（含分页视频列表）。
// GET /api/v1/favorites/:id
// 未登录用户也可查看公开收藏夹。
func (h *Handler) GetFavoriteDetail(c *gin.Context) {
	userID := middleware.GetUserID(c)
	// Allow unauthenticated viewers to see public favorites

	// 解析收藏夹 ID
	favoriteID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "收藏夹ID格式错误")
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))

	resp, err := h.svc.GetFavoriteDetail(c.Request.Context(), userID, uint(favoriteID), page, pageSize)
	if err != nil {
		// 区分服务层错误：私有收藏夹返回 403，不存在返回 404
		var svcErr *favoriteservice.Error
		if errors.As(err, &svcErr) {
			status := http.StatusForbidden
			if svcErr.Code == errcode.FavoriteNotFound {
				status = http.StatusNotFound
			}
			response.Error(c, status, svcErr.Code, svcErr.Msg)
			return
		}
		logger.Error("get favorite detail failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}

// GetUserFavorites 获取指定用户的公开收藏夹列表（用于个人主页展示）。
// GET /api/v1/users/:id/favorites
func (h *Handler) GetUserFavorites(c *gin.Context) {
	// 解析目标用户 ID
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "用户ID格式错误")
		return
	}

	resps, err := h.svc.GetUserPublicFavorites(c.Request.Context(), uint(targetID))
	if err != nil {
		logger.Error("get user favorites failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resps)
}

// ToggleFavoriteItem 切换收藏项：已收藏则取消，未收藏则添加。
// POST /api/v1/favorites/:id/items
// 需要登录认证。
func (h *Handler) ToggleFavoriteItem(c *gin.Context) {
	// 获取当前登录用户 ID
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	// 解析收藏夹 ID
	favoriteID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "收藏夹ID格式错误")
		return
	}

	// 绑定请求体（包含视频 ID）
	var req favoritemodel.FavoriteItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请求参数格式错误")
		return
	}

	resp, err := h.svc.ToggleFavoriteItem(c.Request.Context(), userID, uint(favoriteID), req.VideoID)
	if err != nil {
		// 区分服务层错误：私有收藏夹或不存在
		var svcErr *favoriteservice.Error
		if errors.As(err, &svcErr) {
			status := http.StatusForbidden
			if svcErr.Code == errcode.FavoriteNotFound {
				status = http.StatusNotFound
			}
			response.Error(c, status, svcErr.Code, svcErr.Msg)
			return
		}
		logger.Error("toggle favorite item failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}
