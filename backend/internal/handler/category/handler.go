package category

import (
	"net/http"

	categoryservice "github.com/My-TuDo/B-B/backend/internal/service/category"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	svc *categoryservice.Service
}

func NewHandler(svc *categoryservice.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) List(c *gin.Context) {
	data, err := h.svc.List(c.Request.Context())
	if err != nil {
		logger.Error("list categories failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, data)
}
