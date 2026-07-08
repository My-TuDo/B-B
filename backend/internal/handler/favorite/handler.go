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

type Handler struct {
	svc *favoriteservice.Service
}

func NewHandler(svc *favoriteservice.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateFavorite(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	var req favoritemodel.FavoriteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请求参数格式错误")
		return
	}

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

func (h *Handler) GetFavorites(c *gin.Context) {
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

func (h *Handler) GetFavoriteDetail(c *gin.Context) {
	userID := middleware.GetUserID(c)
	// Allow unauthenticated viewers to see public favorites

	favoriteID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "收藏夹ID格式错误")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))

	resp, err := h.svc.GetFavoriteDetail(c.Request.Context(), userID, uint(favoriteID), page, pageSize)
	if err != nil {
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

func (h *Handler) GetUserFavorites(c *gin.Context) {
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

func (h *Handler) ToggleFavoriteItem(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	favoriteID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "收藏夹ID格式错误")
		return
	}

	var req favoritemodel.FavoriteItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请求参数格式错误")
		return
	}

	resp, err := h.svc.ToggleFavoriteItem(c.Request.Context(), userID, uint(favoriteID), req.VideoID)
	if err != nil {
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
