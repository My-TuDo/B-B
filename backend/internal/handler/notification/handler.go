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

type Handler struct {
	svc *messageservice.Service
}

func NewHandler(svc *messageservice.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetNotifications(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

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

func (h *Handler) ReadAll(c *gin.Context) {
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

func (h *Handler) MarkSingleRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

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
