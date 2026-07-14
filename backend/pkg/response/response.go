// Package response 提供统一的 HTTP JSON 响应格式。
// 所有 API 响应均包含 code、message 字段，可选 data 和 request_id。
// 所有 Handler 必须使用本包的 Success/Error 方法返回数据。
// 提供 Success、Created、Error、ErrorWithRequestID 四个便捷方法。
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 是后端所有接口的通用响应信封。
// 无论成功还是失败，所有接口均以此结构体序列化 JSON 返回。
type Response struct {
	Code      int         `json:"code"`                // HTTP 状态码或业务错误码
	Message   string      `json:"message"`             // 中文提示语
	Data      interface{} `json:"data,omitempty"`      // 业务数据（成功时有值，失败时省略）
	RequestID string      `json:"request_id,omitempty"` // 请求追踪 ID（用于日志关联）
}

// Success 返回 HTTP 200 成功响应。
// data 为业务数据，会序列化到 JSON 的 data 字段。
func Success(c *gin.Context, data interface{}) {
	requestID, _ := c.Get("requestId")
	rid := ""
	if requestID != nil {
		rid = requestID.(string)
	}
	c.JSON(http.StatusOK, Response{
		Code:      200,
		Message:   "成功",
		Data:      data,
		RequestID: rid,
	})
}

// Created 返回 HTTP 201 创建成功响应。
// 用于 POST 请求成功创建资源后的标准响应。
func Created(c *gin.Context, data interface{}) {
	requestID, _ := c.Get("requestId")
	rid := ""
	if requestID != nil {
		rid = requestID.(string)
	}
	c.JSON(http.StatusCreated, Response{
		Code:      201,
		Message:   "创建成功",
		Data:      data,
		RequestID: rid,
	})
}

// Error 返回错误响应，需指定 HTTP 状态码、业务错误码和中文消息。
// httpStatus 为 HTTP 状态码（如 400、401、500）；
// code 为业务错误码（来自 errcode 包）；
// message 为面向用户的中文提示语。
func Error(c *gin.Context, httpStatus int, code int, message string) {
	requestID, _ := c.Get("requestId")
	rid := ""
	if requestID != nil {
		rid = requestID.(string)
	}
	c.JSON(httpStatus, Response{
		Code:      code,
		Message:   message,
		RequestID: rid,
	})
}

// ErrorWithRequestID 返回携带指定 requestID 的错误响应。
// 用于 Recovery 中间件等上下文中 Gin Context 的 requestId 可能丢失的场景。
func ErrorWithRequestID(c *gin.Context, httpStatus int, code int, message string, requestID string) {
	c.JSON(httpStatus, Response{
		Code:      code,
		Message:   message,
		RequestID: requestID,
	})
}
