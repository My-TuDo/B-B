// Package category 提供视频分类相关的 HTTP 处理层。
// 负责解析请求、调用 service 层并返回统一格式的 JSON 响应。
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

// Handler 分类处理器，持有分类服务的引用。
type Handler struct {
	svc *categoryservice.Service
}

// NewHandler 创建分类处理器实例。
func NewHandler(svc *categoryservice.Service) *Handler {
	return &Handler{svc: svc}
}

// List 获取全部分类列表。
// GET /api/v1/categories
func (h *Handler) List(c *gin.Context) {
	data, err := h.svc.List(c.Request.Context())
	if err != nil {
		logger.Error("list categories failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, data)
}
