package search

import (
	"net/http"
	"strconv"

	searchservice "github.com/My-TuDo/B-B/backend/internal/service/search"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	svc *searchservice.Service
}

func NewHandler(svc *searchservice.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Search(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "搜索关键词不能为空")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))

	resp, err := h.svc.Search(c.Request.Context(), q, page, pageSize)
	if err != nil {
		logger.Error("search failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}

func (h *Handler) Suggestions(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		response.Success(c, []interface{}{})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 {
		limit = 10
	}
	if limit > 20 {
		limit = 20
	}

	resp, err := h.svc.Suggestions(c.Request.Context(), q, limit)
	if err != nil {
		logger.Error("search suggestions failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}
