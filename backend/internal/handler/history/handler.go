package history

import (
	"net/http"
	"strconv"

	"github.com/My-TuDo/B-B/backend/internal/middleware"
	historymodel "github.com/My-TuDo/B-B/backend/internal/model/history"
	historyservice "github.com/My-TuDo/B-B/backend/internal/service/history"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	svc *historyservice.Service
}

func NewHandler(svc *historyservice.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateOrUpdate(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	var req historymodel.CreateHistoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请求参数格式错误")
		return
	}

	if err := h.svc.CreateOrUpdate(c.Request.Context(), userID, &req); err != nil {
		logger.Error("create/update history failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, nil)
}

func (h *Handler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))

	resp, err := h.svc.List(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		logger.Error("list history failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}
