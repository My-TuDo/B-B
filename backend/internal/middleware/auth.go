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

var rdbClient *redis.Client

func InitAuth(rdb *redis.Client) {
	rdbClient = rdb
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read token from Cookie
		token, err := c.Cookie("token")
		if err != nil || token == "" {
			// Also try Authorization header
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
			c.Abort()
			return
		}

		claims, err := jwt.ParseToken(token)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, errcode.TokenInvalid, errcode.Message(errcode.TokenInvalid))
			c.Abort()
			return
		}

		// Redis whitelist check
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

		c.Set("userId", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// GetUserID is a helper for handlers to extract userId.
func GetUserID(c *gin.Context) uint {
	val, exists := c.Get("userId")
	if !exists {
		return 0
	}
	return val.(uint)
}

// GetUsername is a helper for handlers to extract username.
func GetUsername(c *gin.Context) string {
	val, exists := c.Get("username")
	if !exists {
		return ""
	}
	return val.(string)
}

// GetRole is a helper for handlers to extract role.
func GetRole(c *gin.Context) uint8 {
	val, exists := c.Get("role")
	if !exists {
		return 0
	}
	return val.(uint8)
}
