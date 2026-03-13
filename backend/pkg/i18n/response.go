package i18n

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response unified response structure
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// SuccessResponse success response
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: GetLocalizedMessageWithGin(c, CodeSuccess),
		Data:    data,
	})
}

// ErrorResponse 错误响应
func ErrorResponse(c *gin.Context, code int, message string) {
	if message == "" {
		message = GetLocalizedMessageWithGin(c, code)
	}

	status := http.StatusInternalServerError
	switch {
	case code == CodeSuccess:
		status = http.StatusOK
	case code == CodeBadRequest || (code >= 1000 && code < 1001) || (code >= 7000 && code < 7100) || (code >= 9100 && code < 9200):
		status = http.StatusBadRequest
	case code == CodeUnauthorized || code == CodeInvalidToken || code == CodeTokenExpired || code == CodeMissingToken || (code >= 8100 && code < 8114):
		status = http.StatusUnauthorized
	case code == CodeForbidden || code == CodeInsufficientPermissions || code == CodeAccessDenied || (code >= 3000 && code < 3100):
		status = http.StatusForbidden
	case code == CodeNotFound || code == CodeResourceNotFound || (code >= 1003 && code < 1004) || code == CodeUserNotFound:
		status = http.StatusNotFound
	case code == CodeMethodNotAllowed:
		status = http.StatusMethodNotAllowed
	case code == CodeTooManyRequests:
		status = http.StatusTooManyRequests
	}

	c.JSON(status, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// ErrorWithCode 错误响应
func ErrorWithCode(c *gin.Context, code int) {
	ErrorResponse(c, code, "")
}

// ErrorWithData 带数据的错误响应
func ErrorWithData(c *gin.Context, code int, message string, data interface{}) {
	if message == "" {
		message = GetLocalizedMessageWithGin(c, code)
	}
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// ErrorResponseWithArgs 带参数的错误响应
func ErrorResponseWithArgs(c *gin.Context, code int, args ...interface{}) {
	message := GetLocalizedMessageWithGin(c, code, args...)
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// BadRequest 400 错误
func BadRequest(c *gin.Context, message string) {
	if message == "" {
		message = GetLocalizedMessageWithGin(c, CodeBadRequest)
	}
	c.JSON(http.StatusBadRequest, Response{
		Code:    CodeBadRequest,
		Message: message,
		Data:    nil,
	})
}

// Unauthorized 401 错误
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = GetLocalizedMessageWithGin(c, CodeUnauthorized)
	}
	c.JSON(http.StatusUnauthorized, Response{
		Code:    CodeUnauthorized,
		Message: message,
		Data:    nil,
	})
}

// Forbidden 403 错误
func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = GetLocalizedMessageWithGin(c, CodeForbidden)
	}
	c.JSON(http.StatusForbidden, Response{
		Code:    CodeForbidden,
		Message: message,
		Data:    nil,
	})
}

// NotFound 404 错误
func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = GetLocalizedMessageWithGin(c, CodeNotFound)
	}
	c.JSON(http.StatusNotFound, Response{
		Code:    CodeNotFound,
		Message: message,
		Data:    nil,
	})
}

// InternalServerError 500 错误
func InternalServerError(c *gin.Context, message string) {
	if message == "" {
		message = GetLocalizedMessageWithGin(c, CodeInternalError)
	}
	c.JSON(http.StatusInternalServerError, Response{
		Code:    CodeInternalError,
		Message: message,
		Data:    nil,
	})
}

// ServiceUnavailable 503 错误
func ServiceUnavailable(c *gin.Context, message string) {
	if message == "" {
		message = GetLocalizedMessageWithGin(c, CodeServiceUnavailable)
	}
	c.JSON(http.StatusServiceUnavailable, Response{
		Code:    CodeServiceUnavailable,
		Message: message,
		Data:    nil,
	})
}

// GatewayTimeout 504 错误
func GatewayTimeout(c *gin.Context, message string) {
	if message == "" {
		message = GetLocalizedMessageWithGin(c, CodeGatewayTimeout)
	}
	c.JSON(http.StatusGatewayTimeout, Response{
		Code:    CodeGatewayTimeout,
		Message: message,
		Data:    nil,
	})
}

// HandleGinError 处理 Gin 错误
func HandleGinError(c *gin.Context, err error) {
	InternalServerError(c, err.Error())
}

// HandleValidationError 处理验证错误
func HandleValidationError(c *gin.Context, message string) {
	ErrorResponse(c, CodeDataValidation, message)
}

// HandleAuthError 处理认证错误
func HandleAuthError(c *gin.Context, message string) {
	Unauthorized(c, message)
}

// HandlePermissionError 处理权限错误
func HandlePermissionError(c *gin.Context, message string) {
	Forbidden(c, message)
}

// HandleSignatureError 处理签名错误
func HandleSignatureError(c *gin.Context, message string) {
	BadRequest(c, message)
}
