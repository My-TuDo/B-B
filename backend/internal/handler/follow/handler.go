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

type Handler struct {
	svc            *followservice.Service
	interactionSvc interface{} // placeholder, used from main
}

func NewHandler(svc *followservice.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ToggleFollow(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	targetID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "用户ID格式错误")
		return
	}

	resp, err := h.svc.ToggleFollow(c.Request.Context(), userID, uint(targetID))
	if err != nil {
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

func (h *Handler) GetFollowers(c *gin.Context) {
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "用户ID格式错误")
		return
	}

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

func (h *Handler) GetFollowing(c *gin.Context) {
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "用户ID格式错误")
		return
	}

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

func (h *Handler) GetFeed(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

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

func (h *Handler) GetProfile(c *gin.Context) {
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "用户ID格式错误")
		return
	}

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
