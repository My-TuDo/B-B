package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// CSRFConfig CSRF 防护配置。
type CSRFConfig struct {
	// PublicPaths 跳过 CSRF 检查的路径列表（GET/HEAD/OPTIONS 已自动跳过）。
	// 用于登录、注册等公开写接口。
	PublicPaths []string
}

// defaultPublicPaths 默认公开路径：注册、登录、刷新 Token。
var defaultPublicPaths = []string{
	"/api/v1/auth/register",
	"/api/v1/auth/login",
	"/api/v1/auth/refresh",
}

// CSRF 返回 CSRF 防护中间件。
//
// 防护策略（Double Submit Cookie）：
//  1. GET/HEAD/OPTIONS 请求直接放行
//  2. 公开路径（注册/登录）直接放行
//  3. 其他写请求必须携带 X-CSRF-Token 头，且与 csrf_token Cookie 值一致
func CSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		// GET、HEAD、OPTIONS 不检查 CSRF
		method := c.Request.Method
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			c.Next()
			return
		}

		// 公开路径跳过检查
		path := c.Request.URL.Path
		for _, pp := range defaultPublicPaths {
			if path == pp {
				c.Next()
				return
			}
		}

		// 检查 X-CSRF-Token 请求头
		csrfHeader := c.GetHeader("X-CSRF-Token")
		if csrfHeader == "" {
			response.Error(c, http.StatusForbidden, errcode.Forbidden, "CSRF token缺失")
			c.Abort()
			return
		}

		// 检查 csrf_token Cookie
		csrfCookie, err := c.Cookie("csrf_token")
		if err != nil || csrfCookie == "" {
			response.Error(c, http.StatusForbidden, errcode.Forbidden, "CSRF cookie缺失")
			c.Abort()
			return
		}

		// 比较 Header 和 Cookie 中的 Token（大小写不敏感）
		if !strings.EqualFold(csrfHeader, csrfCookie) {
			response.Error(c, http.StatusForbidden, errcode.Forbidden, "CSRF token不匹配")
			c.Abort()
			return
		}

		c.Next()
	}
}

// GenerateCSRFToken 生成 32 字节随机数的十六进制字符串。
func GenerateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// SetCSRFCookie 设置 csrf_token Cookie 并返回 Token 值。
// Cookie 属性：SameSite=Lax，有效期 24 小时，HttpOnly=false（前端需读取）。
func SetCSRFCookie(c *gin.Context) string {
	token, err := GenerateCSRFToken()
	if err != nil {
		// 生成失败时使用降级方案
		token = hex.EncodeToString(make([]byte, 32))
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("csrf_token", token, 86400, "/", "", false, false)
	return token
}
