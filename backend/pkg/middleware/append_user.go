package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/kymo-mcp/mcpcan/pkg/logger"
)

// AppendUserMiddleware extracts X-Consum-User-Id from header and sets it in context
func AppendUserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract X-Consum-User-Id from header
		userIDStr := c.GetHeader("X-Consum-User-Id")
		
		if userIDStr != "" {
			userID, err := strconv.ParseInt(userIDStr, 10, 64)
			if err != nil {
				logger.Warn("Failed to parse X-Consum-User-Id", zap.String("userIDStr", userIDStr), zap.Error(err))
			} else {
				// Set userId in context if valid
				c.Set("userId", userID)
			}
		}

		c.Next()
	}
}
