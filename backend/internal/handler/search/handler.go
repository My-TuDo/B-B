// Package search 提供搜索相关的 HTTP 处理层。
// 处理视频搜索和搜索建议功能。
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

// Handler 搜索处理器，持有搜索服务的引用。
type Handler struct {
	svc *searchservice.Service
}

// NewHandler 创建搜索处理器实例。
func NewHandler(svc *searchservice.Service) *Handler {
	return &Handler{svc: svc}
}

// Search 根据关键词搜索视频（分页）。
// GET /api/v1/search?q=关键词&page=1&page_size=12
func (h *Handler) Search(c *gin.Context) {
	// 获取搜索关键词
	q := c.Query("q")
	if q == "" {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "搜索关键词不能为空")
		return
	}

	// 解析分页参数
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

// Suggestions 获取搜索建议（自动补全）。
// GET /api/v1/search/suggestions?q=关键词&limit=10
// limit 范围 1-20，默认 10。
func (h *Handler) Suggestions(c *gin.Context) {
	// 获取搜索前缀
	q := c.Query("q")
	if q == "" {
		response.Success(c, []interface{}{})
		return
	}

	// 解析建议数量限制，限制在 [1, 20] 范围内
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
