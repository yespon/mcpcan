package middleware

import (
	"github.com/kymo-mcp/mcpcan/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// ErrorHandler Error handling middleware
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Only process if there are errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			// Return corresponding error code based on error type
			switch err.Type {
			case gin.ErrorTypeBind:
				// Request parameter binding error
				i18n.BadRequest(c, "Invalid request parameters")
			case gin.ErrorTypeRender:
				// Rendering error
				i18n.InternalServerError(c, "Server render error")
			case gin.ErrorTypePrivate:
				// Private error
				i18n.InternalServerError(c, "Internal server error")
			case gin.ErrorTypePublic:
				// Public error
				i18n.BadRequest(c, err.Error())
			default:
				// Default internal server error
				i18n.InternalServerError(c, err.Error())
			}
		}
	}
}

// NotFoundHandler 404 Handler
func NotFoundHandler(c *gin.Context) {
	i18n.NotFound(c, "Resource not found")
}

// MethodNotAllowedHandler 405 Handler
func MethodNotAllowedHandler(c *gin.Context) {
	i18n.ErrorResponse(c, i18n.CodeMethodNotAllowed, "Method not allowed")
}

// PanicRecovery Panic recovery middleware
func PanicRecovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		i18n.InternalServerError(c, "Internal server error")
	})
}
