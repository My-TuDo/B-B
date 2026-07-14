package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS 处理跨域请求。
// 允许来源：localhost:3000（前端开发服务器）。
// 允许方法：GET、POST、PUT、DELETE、OPTIONS。
// 允许携带 Cookie（Access-Control-Allow-Credentials）。
// 预检请求缓存 24 小时。
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置 CORS 响应头
		c.Header("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization,X-Request-Id")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 预检缓存 24 小时

		// OPTIONS 预检请求直接返回 204
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
