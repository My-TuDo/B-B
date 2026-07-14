// Package middleware 提供 Gin HTTP 中间件集合。
// 包含：Recovery（Panic 恢复）、RequestID（请求追踪）、Logger（请求日志）、
// CORS（跨域）、RateLimit（限流）、CSRF（跨站请求伪造防护）、Auth（JWT 认证）。
package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/jwt"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// rdbClient 认证中间件使用的 Redis 客户端，由 InitAuth 注入。
// 用于 Token 白名单校验，防止已撤销的 Token 继续使用。
var rdbClient *redis.Client

// InitAuth 初始化认证模块所需的 Redis 客户端。
// 必须在服务启动时调用（main.go 第 4 步），否则 Token 白名单校验将被跳过。
func InitAuth(rdb *redis.Client) {
	rdbClient = rdb
}

// AuthRequired 返回 JWT 认证中间件。
// 将该中间件挂载到需要登录保护的路由组上。
//
// 认证流程：
//  1. 优先从 Cookie 读取 Token，其次从 Authorization: Bearer <token> 头读取
//  2. 解析并验证 JWT（签名 + 有效期）
//  3. Redis 白名单校验（确保 Token 未被撤销）
//  4. 将 userId、username、role 注入 Gin Context，供后续 Handler 使用
//
// 任一环节失败均返回 401 并中断请求链。
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 步骤 1：从 Cookie 读取 Token
		token, err := c.Cookie("token")
		if err != nil || token == "" {
			// 尝试从 Authorization 头读取
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && parts[0] == "Bearer" {
					token = parts[1]
				}
			}
		}

		// 未找到 Token
		if token == "" {
			response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
			c.Abort()
			return
		}

		// 步骤 2：解析 JWT
		claims, err := jwt.ParseToken(token)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, errcode.TokenInvalid, errcode.Message(errcode.TokenInvalid))
			c.Abort()
			return
		}

		// 步骤 3：Redis 白名单校验（防止 Token 被撤销后仍可使用）
		if rdbClient != nil {
			ctx := c.Request.Context()
			key := fmt.Sprintf("auth:token:%d", claims.UserID)
			storedToken, err := rdbClient.Get(ctx, key).Result()
			if err != nil || storedToken != token {
				response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
				c.Abort()
				return
			}
		}

		// 步骤 4：注入用户信息到上下文
		c.Set("userId", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// GetUserID 从 Gin 上下文提取已验证的用户 ID。
// 仅在 AuthRequired 中间件之后调用有效；未认证时返回 0。
func GetUserID(c *gin.Context) uint {
	val, exists := c.Get("userId")
	if !exists {
		return 0
	}
	return val.(uint)
}

// GetUsername 从 Gin 上下文提取已验证的用户名。
// 仅在 AuthRequired 中间件之后调用有效；未认证时返回空字符串。
func GetUsername(c *gin.Context) string {
	val, exists := c.Get("username")
	if !exists {
		return ""
	}
	return val.(string)
}

// GetRole 从 Gin 上下文提取已验证的用户角色。
// 0 = 普通用户，1 = 管理员。
// 仅在 AuthRequired 中间件之后调用有效；未认证时返回 0。
func GetRole(c *gin.Context) uint8 {
	val, exists := c.Get("role")
	if !exists {
		return 0
	}
	return val.(uint8)
}
