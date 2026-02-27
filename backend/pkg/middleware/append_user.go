package middleware

import (
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kymo-mcp/mcpcan/api/authz/user_auth"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/i18n"
	"go.uber.org/zap"

	"github.com/kymo-mcp/mcpcan/pkg/gomap"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
)

// AppendUserMiddleware extracts X-Consum-User-Id from header and sets it in context
func AppendUserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			gomap.Del(common.UserInfoContextKey)
		}()

		// Extract X-Consum-User-Id from header
		userIDStr := c.GetHeader(common.UserIdHeaderKey)

		var u user_auth.UserInfo
		if userIDStr != "" {
			userID, err := strconv.ParseInt(userIDStr, 10, 64)
			if err != nil {
				logger.Warn("Failed to parse X-Consum-User-Id", zap.String("userIDStr", userIDStr), zap.Error(err))
			} else {
				// Set userId in context if valid
				c.Set("userId", userID)
				u.UserId = userID
			}
		}

		userInfo := c.Request.Header.Get(common.UserInfoHeaderKey)
		if userInfo != "" {
			userInfoBytes, err := base64.StdEncoding.DecodeString(userInfo)
			if err != nil {
				i18n.Unauthorized(c, "invalid user token")
				c.Abort()
				return
			}
			err = json.Unmarshal(userInfoBytes, &u)
			if err != nil {
				i18n.Unauthorized(c, "invalid user token")
				c.Abort()
				return
			}
		}

		if u.UserId != 0 || len(u.RoleIds) > 0 {
			c.Set(common.UserInfoContextKey, &u)
			gomap.Set(common.UserInfoContextKey, &u)
		}

		c.Next()
	}
}
