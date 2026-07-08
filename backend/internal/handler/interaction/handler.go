package interaction

import (
	"net/http"
	"strconv"

	"github.com/My-TuDo/B-B/backend/internal/middleware"
	interactionservice "github.com/My-TuDo/B-B/backend/internal/service/interaction"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	svc *interactionservice.Service
}

func NewHandler(svc *interactionservice.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetVideoInteractions(c *gin.Context) {
	userID := middleware.GetUserID(c)

	videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "视频ID格式错误")
		return
	}

	status, err := h.svc.GetVideoInteractions(c.Request.Context(), userID, uint(videoID))
	if err != nil {
		logger.Error("get video interactions failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, status)
}
