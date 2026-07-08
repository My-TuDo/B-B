package tag

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/My-TuDo/B-B/backend/internal/middleware"
	tagmodel "github.com/My-TuDo/B-B/backend/internal/model/tag"
	tagservice "github.com/My-TuDo/B-B/backend/internal/service/tag"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	svc *tagservice.Service
}

func NewHandler(svc *tagservice.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Create(c *gin.Context) {
	var req tagmodel.CreateTagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请求参数格式错误")
		return
	}

	name := req.Name
	if name == "" || len([]rune(name)) > 30 {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "标签名不能为空且不超过30字符")
		return
	}

	resp, err := h.svc.Create(c.Request.Context(), name)
	if err != nil {
		logger.Error("create tag failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Created(c, resp)
}

func (h *Handler) List(c *gin.Context) {
	resp, err := h.svc.List(c.Request.Context())
	if err != nil {
		logger.Error("list tags failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}

func (h *Handler) SetVideoTags(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	videoIDStr := c.Param("id")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "视频ID格式错误")
		return
	}

	var req tagmodel.SetVideoTagsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请求参数格式错误")
		return
	}

	if err := h.svc.SetVideoTags(c.Request.Context(), userID, uint(videoID), &req); err != nil {
		var svcErr *tagservice.Error
		if errors.As(err, &svcErr) {
			status := http.StatusNotFound
			if svcErr.Code == errcode.Forbidden {
				status = http.StatusForbidden
			}
			response.Error(c, status, svcErr.Code, svcErr.Msg)
			return
		}
		logger.Error("set video tags failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, nil)
}

func (h *Handler) GetVideoTags(c *gin.Context) {
	videoIDStr := c.Param("id")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "视频ID格式错误")
		return
	}

	resp, err := h.svc.GetVideoTags(c.Request.Context(), uint(videoID))
	if err != nil {
		logger.Error("get video tags failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}
