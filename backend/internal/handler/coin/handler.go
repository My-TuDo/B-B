// Package coin 提供投币相关的 HTTP 处理层。
// 处理用户对视频的投币操作，支持投 1 或 2 个币。
package coin

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/My-TuDo/B-B/backend/internal/middleware"
	coinmodel "github.com/My-TuDo/B-B/backend/internal/model/coin"
	coinservice "github.com/My-TuDo/B-B/backend/internal/service/coin"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler 投币处理器，持有投币服务的引用。
type Handler struct {
	svc *coinservice.Service
}

// NewHandler 创建投币处理器实例。
func NewHandler(svc *coinservice.Service) *Handler {
	return &Handler{svc: svc}
}

// AddCoin 对指定视频进行投币操作。
// POST /api/v1/videos/:id/coin
// 需要登录认证，投币数量只能为 1 或 2。
func (h *Handler) AddCoin(c *gin.Context) {
	// 获取当前登录用户 ID
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	// 解析视频 ID
	videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "视频ID格式错误")
		return
	}

	// 绑定请求体
	var req coinmodel.CoinReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请求参数格式错误")
		return
	}

	// 校验投币数量：只允许 1 或 2
	if req.Count != 1 && req.Count != 2 {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "投币数量只能为1或2")
		return
	}

	// 执行投币逻辑
	resp, err := h.svc.AddCoin(c.Request.Context(), userID, uint(videoID), req.Count)
	if err != nil {
		var svcErr *coinservice.Error
		if errors.As(err, &svcErr) {
			response.Error(c, http.StatusBadRequest, svcErr.Code, svcErr.Msg)
			return
		}
		logger.Error("add coin failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}
