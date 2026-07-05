package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

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

func ErrorWithRequestID(c *gin.Context, httpStatus int, code int, message string, requestID string) {
	c.JSON(httpStatus, Response{
		Code:      code,
		Message:   message,
		RequestID: requestID,
	})
}
