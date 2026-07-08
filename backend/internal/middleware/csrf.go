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

type CSRFConfig struct {
	// PublicPaths are paths that skip CSRF check (GET/HEAD/OPTIONS already skipped).
	// These are for public write routes like login/register.
	PublicPaths []string
}

var defaultPublicPaths = []string{
	"/api/v1/auth/register",
	"/api/v1/auth/login",
	"/api/v1/auth/refresh",
}

func CSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip GET, HEAD, OPTIONS
		method := c.Request.Method
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			c.Next()
			return
		}

		// Skip public paths
		path := c.Request.URL.Path
		for _, pp := range defaultPublicPaths {
			if path == pp {
				c.Next()
				return
			}
		}

		// Check CSRF header vs cookie
		csrfHeader := c.GetHeader("X-CSRF-Token")
		if csrfHeader == "" {
			response.Error(c, http.StatusForbidden, errcode.Forbidden, "CSRF token缺失")
			c.Abort()
			return
		}

		csrfCookie, err := c.Cookie("csrf_token")
		if err != nil || csrfCookie == "" {
			response.Error(c, http.StatusForbidden, errcode.Forbidden, "CSRF cookie缺失")
			c.Abort()
			return
		}

		if !strings.EqualFold(csrfHeader, csrfCookie) {
			response.Error(c, http.StatusForbidden, errcode.Forbidden, "CSRF token不匹配")
			c.Abort()
			return
		}

		c.Next()
	}
}

// GenerateCSRFToken generates a random 32-byte hex string.
func GenerateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// SetCSRFCookie sets the csrf_token cookie and returns the token value.
func SetCSRFCookie(c *gin.Context) string {
	token, err := GenerateCSRFToken()
	if err != nil {
		token = hex.EncodeToString(make([]byte, 32)) // fallback
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("csrf_token", token, 86400, "/", "", false, false)
	return token
}
