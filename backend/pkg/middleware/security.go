package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
)

// SecurityConfig Security configuration
type SecurityConfig struct {
	SecretKey    string        // Signature secret key
	ReplayWindow time.Duration // Replay attack time window
	EnableReplay bool          // Whether to enable anti-replay
	EnableSign   bool          // Whether to enable anti-tamper
}

// SecurityMiddleware Security middleware, implementing anti-tamper and anti-replay attack
func SecurityMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip security check for mcp-gateway prefix
		if strings.HasPrefix(c.Request.URL.Path, common.GatewayRoutePrefix) {
			c.Next()
			return
		}

		config := &SecurityConfig{
			SecretKey:    secret,
			ReplayWindow: common.ReplayWindow,
			EnableReplay: common.EnableReplay,
			EnableSign:   common.EnableSign,
		}

		// Anti-replay attack check
		if config.EnableReplay {
			if err := checkReplayAttack(c, config.ReplayWindow); err != nil {
				logger.Warn("Replay check failed", zap.Error(err), zap.String("path", c.Request.URL.Path))
				i18n.HandleSignatureError(c, "Request expired or repeated")
				c.Abort()
				return
			}
		}

		// Anti-tamper check
		if config.EnableSign {
			if err := checkSignature(c, config.SecretKey); err != nil {
				logger.Warn("Signature verification failed", zap.Error(err), zap.String("path", c.Request.URL.Path))
				i18n.HandleSignatureError(c, "Signature verification failed")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// checkReplayAttack Check replay attack
func checkReplayAttack(c *gin.Context, window time.Duration) error {
	timestampStr := c.GetHeader("X-Timestamp")
	if timestampStr == "" {
		return fmt.Errorf("missing timestamp")
	}

	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp format")
	}

	requestTime := time.Unix(timestamp, 0)
	now := time.Now()

	// Check time window
	if now.Sub(requestTime) > window {
		return fmt.Errorf("request expired")
	}

	// Check if time is too far ahead
	if requestTime.Sub(now) > time.Minute {
		return fmt.Errorf("request time too far ahead")
	}

	// TODO: Add nonce check here to prevent duplicate requests with same timestamp
	// Can use Redis or memory cache to store used nonce
	// redisClient := redis.GetClient()
	// nonceKey := fmt.Sprintf("nonce:%d:%s", timestamp, c.Request.URL.Path)

	// // Check if nonce already exists
	// if _, err := redisClient.Get(nonceKey); err == nil {
	// 	return fmt.Errorf("duplicate request")
	// }

	// // Cache nonce, expiration time is window length
	// if err := redisClient.Set(nonceKey, "used", window); err != nil {
	// 	return fmt.Errorf("failed to cache nonce: %v", err)
	// }
	return nil
}

// checkSignature Check signature
func checkSignature(c *gin.Context, secretKey string) error {
	signature := c.GetHeader("X-Signature")
	if signature == "" {
		return fmt.Errorf("missing signature")
	}

	// Build sign string
	signString, err := buildSignString(c)
	if err != nil {
		return fmt.Errorf("failed to build sign string: %v", err)
	}

	// Calculate expected signature
	expectedSignature := calculateSignature(signString, secretKey)

	// Verify signature
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return fmt.Errorf("signature mismatch")
	}

	return nil
}

// buildSignString Build sign string
func buildSignString(c *gin.Context) (string, error) {
	var parts []string

	// Add HTTP method
	parts = append(parts, c.Request.Method)

	// Add path
	parts = append(parts, c.Request.URL.Path)

	// Add timestamp
	timestamp := c.GetHeader("X-Timestamp")
	if timestamp != "" {
		parts = append(parts, timestamp)
	}

	// Add query parameters (sorted alphabetically)
	if len(c.Request.URL.RawQuery) > 0 {
		queryParams := make([]string, 0)
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				queryParams = append(queryParams, fmt.Sprintf("%s=%s", key, value))
			}
		}
		sort.Strings(queryParams)
		parts = append(parts, strings.Join(queryParams, "&"))
	}

	// Add request body (if POST/PUT methods)
	if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
		body := c.GetHeader("X-Body-Hash")
		if body != "" {
			parts = append(parts, body)
		}
	}

	return strings.Join(parts, "|"), nil
}

// calculateSignature Calculate signature
func calculateSignature(data, secretKey string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
