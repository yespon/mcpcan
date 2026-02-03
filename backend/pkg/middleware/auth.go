package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/kymo-mcp/mcpcan/pkg/jwt"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
	"github.com/kymo-mcp/mcpcan/pkg/redis"
)

var SkipPaths = []string{
	"/health",
	"/authz/encryption-key",
	"/authz/login",
	"/authz/logout",
	"/authz/register",
	"/authz/refresh",
	"/authz/validate",
	"/market/code/download",
	"/market/openapi/download",
}

// AuthTokenMiddleware User token validation middleware
func AuthTokenMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := ExtractToken(c)
		if tokenString == "" {
			i18n.Unauthorized(c, "missing auth token")
			c.Abort()
			return
		}

		claims, err := jwt.ParseTokenWithClaims(tokenString, secret)
		if err != nil {
			logger.Error("JWT token validation failed", zap.Error(err))
			i18n.Unauthorized(c, "invalid auth token")
			c.Abort()
			return
		}

		// Check if token is expired
		if time.Now().Unix() > int64(claims.ExpiresAt.Unix()) {
			i18n.Unauthorized(c, "auth token expired")
			c.Abort()
			return
		}

		if claims.UserID == 0 || len(claims.Username) == 0 {
			i18n.Unauthorized(c, "user id or username is empty")
			c.Abort()
			return
		}

		userToken, err := redis.GetUserTokenByToken(tokenString)
		if err != nil {
			logger.Error("failed to get user token", zap.Error(err))
			i18n.Unauthorized(c, "invalid auth token")
			c.Abort()
			return
		}
		if userToken == nil || userToken.UserID != uint(claims.UserID) {
			i18n.Unauthorized(c, "user id not match")
			c.Abort()
			return
		}

		// Check if token is valid
		c.Set("userId", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

// ExtractToken Extract token
func ExtractToken(c *gin.Context) string {
	// Extract from Authorization header
	auth := c.GetHeader("Authorization")
	if auth != "" && strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	// Extract from query parameter
	token := c.Query("token")
	if token != "" {
		return token
	}

	return ""
}
