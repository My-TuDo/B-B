package user

import (
	"errors"
	"net/http"
	"strconv"

	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	userservice "github.com/My-TuDo/B-B/backend/internal/service/user"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	svc *userservice.Service
}

func NewHandler(svc *userservice.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "用户ID格式错误")
		return
	}

	resp, err := h.svc.GetUser(c.Request.Context(), uint(id))
	if err != nil {
		var svcErr *userservice.Error
		if errors.As(err, &svcErr) {
			response.Error(c, http.StatusNotFound, svcErr.Code, svcErr.Msg)
			return
		}
		logger.Error("get user failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}

func (h *Handler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "用户ID格式错误")
		return
	}

	// User isolation: only the user themselves can update
	userID := middleware.GetUserID(c)
	if uint(id) != userID {
		response.Error(c, http.StatusForbidden, errcode.Forbidden, errcode.Message(errcode.Forbidden))
		return
	}

	var req usermodel.UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请求参数格式错误")
		return
	}

	resp, err := h.svc.UpdateUser(c.Request.Context(), uint(id), &req)
	if err != nil {
		var svcErr *userservice.Error
		if errors.As(err, &svcErr) {
			response.Error(c, http.StatusNotFound, svcErr.Code, svcErr.Msg)
			return
		}
		logger.Error("update user failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}
