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

type Handler struct {
	svc *coinservice.Service
}

func NewHandler(svc *coinservice.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) AddCoin(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "视频ID格式错误")
		return
	}

	var req coinmodel.CoinReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请求参数格式错误")
		return
	}

	if req.Count != 1 && req.Count != 2 {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "投币数量只能为1或2")
		return
	}

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
