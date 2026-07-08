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

type Handler struct {
	svc *creatorservice.Service
}

func NewHandler(svc *creatorservice.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreatorVideos(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	role := middleware.GetRole(c)
	if role < 2 {
		response.Error(c, http.StatusForbidden, errcode.Forbidden, errcode.Message(errcode.Forbidden))
		return
	}

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

func (h *Handler) CreatorStats(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

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
