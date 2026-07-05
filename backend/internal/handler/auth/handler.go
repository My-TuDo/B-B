package auth

import (
	"errors"
	"net/http"
	"strings"

	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
	authservice "github.com/My-TuDo/B-B/backend/internal/service/auth"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/My-TuDo/B-B/backend/pkg/validator"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	svc *authservice.Service
}

func NewHandler(svc *authservice.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(c *gin.Context) {
	var req usermodel.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请求参数格式错误")
		return
	}

	if err := validator.Struct(req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.InvalidParams, formatValidationErr(err))
		return
	}

	resp, token, err := h.svc.Register(c.Request.Context(), &req)
	if err != nil {
		var svcErr *authservice.Error
		if errors.As(err, &svcErr) {
			response.Error(c, http.StatusConflict, svcErr.Code, svcErr.Msg)
			return
		}
		logger.Error("register failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	setCookie(c, token)
	response.Created(c, resp)
}

func (h *Handler) Login(c *gin.Context) {
	var req usermodel.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请求参数格式错误")
		return
	}

	if req.Account == "" || req.Password == "" {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "账号和密码不能为空")
		return
	}

	resp, token, err := h.svc.Login(c.Request.Context(), req.Account, req.Password)
	if err != nil {
		var svcErr *authservice.Error
		if errors.As(err, &svcErr) {
			status := http.StatusBadRequest
			if svcErr.Code == errcode.PasswordWrong {
				status = http.StatusUnauthorized
			} else if svcErr.Code == errcode.UserNotFound {
				status = http.StatusNotFound
			}
			response.Error(c, status, svcErr.Code, svcErr.Msg)
			return
		}
		logger.Error("login failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	setCookie(c, token)
	response.Success(c, resp)
}

func (h *Handler) Logout(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	if err := h.svc.Logout(c.Request.Context(), userID.(uint)); err != nil {
		logger.Error("logout failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	setCookie(c, "")
	response.Success(c, nil)
}

func (h *Handler) Me(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	resp, err := h.svc.GetMe(c.Request.Context(), userID.(uint))
	if err != nil {
		var svcErr *authservice.Error
		if errors.As(err, &svcErr) {
			response.Error(c, http.StatusNotFound, svcErr.Code, svcErr.Msg)
			return
		}
		logger.Error("get me failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}

func (h *Handler) Refresh(c *gin.Context) {
	token, err := c.Cookie("token")
	if err != nil || token == "" {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				token = parts[1]
			}
		}
	}

	if token == "" {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	newToken, err := h.svc.Refresh(c.Request.Context(), token)
	if err != nil {
		var svcErr *authservice.Error
		if errors.As(err, &svcErr) {
			response.Error(c, http.StatusUnauthorized, svcErr.Code, svcErr.Msg)
			return
		}
		logger.Error("refresh token failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	setCookie(c, newToken)
	response.Success(c, nil)
}

func setCookie(c *gin.Context, token string) {
	maxAge := 604800 // 7 days
	if token == "" {
		maxAge = -1
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("token", token, maxAge, "/", "", false, true)
}

func formatValidationErr(err error) string {
	return err.Error()
}
