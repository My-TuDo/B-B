package danmaku

import (
	"net/http"
	"strconv"

	"github.com/My-TuDo/B-B/backend/internal/middleware"
	danmakumodel "github.com/My-TuDo/B-B/backend/internal/model/danmaku"
	danmakuservice "github.com/My-TuDo/B-B/backend/internal/service/danmaku"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	svc *danmakuservice.Service
}

func NewHandler(svc *danmakuservice.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetDanmaku(c *gin.Context) {
	videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "视频ID格式错误")
		return
	}

	resps, err := h.svc.GetDanmaku(c.Request.Context(), uint(videoID))
	if err != nil {
		logger.Error("get danmaku failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resps)
}

func (h *Handler) SendDanmaku(c *gin.Context) {
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

	var req danmakumodel.DanmakuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请求参数格式错误")
		return
	}

	if req.Content == "" || len(req.Content) > 200 {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "弹幕内容1-200字符")
		return
	}

	resp, err := h.svc.SendDanmaku(c.Request.Context(), uint(videoID), userID, &req)
	if err != nil {
		logger.Error("send danmaku failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Created(c, resp)
}
