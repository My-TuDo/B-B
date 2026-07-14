// Package auth 提供认证相关的 HTTP 处理器，包括用户注册、登录、登出、
// 个人信息获取、Token 刷新以及 CSRF 令牌发放等功能。
package auth

import (
	"errors"
	"net/http"
	"strings"

	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
	authservice "github.com/My-TuDo/B-B/backend/internal/service/auth"
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/My-TuDo/B-B/backend/pkg/validator"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler 是认证模块的 HTTP 处理器，持有认证服务实例。
type Handler struct {
	svc *authservice.Service
}

// NewHandler 创建认证处理器实例。
func NewHandler(svc *authservice.Service) *Handler {
	return &Handler{svc: svc}
}

// Register 处理用户注册请求。
// 接收 JSON 格式的注册信息，验证后调用服务层注册，成功时设置 Token Cookie。
func (h *Handler) Register(c *gin.Context) {
	var req usermodel.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请求参数格式错误")
		return
	}

	// 结构体校验（如邮箱格式、密码强度等）
	if err := validator.Struct(req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.InvalidParams, formatValidationErr(err))
		return
	}

	resp, token, err := h.svc.Register(c.Request.Context(), &req)
	if err != nil {
		// 判断是否为业务层错误（如用户名已存在）
		var svcErr *authservice.Error
		if errors.As(err, &svcErr) {
			response.Error(c, http.StatusConflict, svcErr.Code, svcErr.Msg)
			return
		}
		logger.Error("register failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	// 注册成功后设置 Token Cookie
	setCookie(c, token)
	response.Created(c, resp)
}

// Login 处理用户登录请求。
// 支持账号/邮箱 + 密码登录，成功后返回用户信息和 Token。
func (h *Handler) Login(c *gin.Context) {
	var req usermodel.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请求参数格式错误")
		return
	}

	// 基础参数校验：账号和密码不能为空
	if req.Account == "" || req.Password == "" {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "账号和密码不能为空")
		return
	}

	resp, token, err := h.svc.Login(c.Request.Context(), req.Account, req.Password)
	if err != nil {
		var svcErr *authservice.Error
		if errors.As(err, &svcErr) {
			// 根据错误码返回不同 HTTP 状态码
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

	// 登录成功后设置 Token Cookie
	setCookie(c, token)
	response.Success(c, resp)
}

// Logout 处理用户登出请求（需认证）。
// 清除服务端 Redis 中的 Token 并清空客户端 Cookie。
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

	// 清空客户端 Cookie
	setCookie(c, "")
	response.Success(c, nil)
}

// Me 获取当前登录用户的个人信息（需认证）。
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

// Refresh 通过现有 Token 换取新 Token。
// 优先从 Cookie 读取，其次从 Authorization 头中的 Bearer Token 读取。
func (h *Handler) Refresh(c *gin.Context) {
	// 尝试从 Cookie 获取 Token
	token, err := c.Cookie("token")
	if err != nil || token == "" {
		// 回退：从 Authorization 头解析 Bearer Token
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

	// 设置新的 Token Cookie
	setCookie(c, newToken)
	response.Success(c, nil)
}

// CSRF 生成并返回一个 CSRF 令牌，用于防止跨站请求伪造攻击。
func (h *Handler) CSRF(c *gin.Context) {
	token := middleware.SetCSRFCookie(c)
	response.Success(c, gin.H{"token": token})
}

// setCookie 设置或清除名为 "token" 的 HttpOnly Cookie。
// 若 token 为空字符串，则清除 Cookie（maxAge = -1）；否则设置 7 天有效期。
func setCookie(c *gin.Context, token string) {
	maxAge := 604800 // 7 天过期
	if token == "" {
		maxAge = -1 // 清除 Cookie
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("token", token, maxAge, "/", "", false, true)
}

// formatValidationErr 格式化参数校验错误信息为中文提示。
func formatValidationErr(err error) string {
	return "请求参数验证失败"
}
