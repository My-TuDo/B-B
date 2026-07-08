package admin

import (
	"net/http"
	"strconv"

	"github.com/My-TuDo/B-B/backend/internal/middleware"
	historymodel "github.com/My-TuDo/B-B/backend/internal/model/history"
	videorepo "github.com/My-TuDo/B-B/backend/internal/repository/video"
	adminservice "github.com/My-TuDo/B-B/backend/internal/service/admin"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	repo *videorepo.Repository
	svc  *adminservice.Service
}

func NewHandler(repo *videorepo.Repository, svc *adminservice.Service) *Handler {
	return &Handler{repo: repo, svc: svc}
}

func (h *Handler) AdminVideos(c *gin.Context) {
	role := middleware.GetRole(c)
	if role < 3 {
		response.Error(c, http.StatusForbidden, errcode.Forbidden, errcode.Message(errcode.Forbidden))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))

	var statusFilter int8 = 2
	statusStr := c.DefaultQuery("status", "2")
	if s, err := strconv.ParseInt(statusStr, 10, 8); err == nil {
		statusFilter = int8(s)
	}

	resp, err := h.svc.ListVideos(c.Request.Context(), statusFilter, page, pageSize)
	if err != nil {
		logger.Error("admin list videos failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}

func (h *Handler) Review(c *gin.Context) {
	role := middleware.GetRole(c)
	if role < 3 {
		response.Error(c, http.StatusForbidden, errcode.Forbidden, errcode.Message(errcode.Forbidden))
		return
	}

	videoIDStr := c.Param("id")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "视频ID格式错误")
		return
	}

	var req historymodel.AdminReviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请求参数格式错误")
		return
	}

	if req.Status != 1 && req.Status != 3 {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "status只能为1或3")
		return
	}

	video, err := h.repo.FindByID(c.Request.Context(), uint(videoID))
	if err != nil {
		logger.Error("admin review find failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}
	if video == nil {
		response.Error(c, http.StatusNotFound, errcode.VideoNotFound, errcode.Message(errcode.VideoNotFound))
		return
	}

	video.Status = req.Status
	if err := h.repo.Update(c.Request.Context(), video); err != nil {
		logger.Error("admin review update failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, nil)
}
