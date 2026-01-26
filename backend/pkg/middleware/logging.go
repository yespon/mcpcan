package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestResponseLoggingMiddleware Detailed request/response logging middleware
func RequestResponseLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record request start time
		start := time.Now()

		// Read request body
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// Reset request body so subsequent handlers can read it
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}
		// Prepare log fields
		logFields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		// Record request headers
		headers := make(map[string]string)
		for k, v := range c.Request.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}
		logFields = append(logFields, zap.Any("headers", headers))

		// Record query parameters
		if c.Request.URL.RawQuery != "" {
			logFields = append(logFields, zap.String("query", c.Request.URL.RawQuery))
		}

		// Record form parameters
		if err := c.Request.ParseForm(); err == nil {
			if len(c.Request.Form) > 0 {
				logFields = append(logFields, zap.Any("form", c.Request.Form))
			}
		}

		// Check if Content-Type is JSON and try to parse request body
		contentType := c.GetHeader("Content-Type")
		if strings.Contains(contentType, "application/json") && len(requestBody) > 0 {
			var jsonBody interface{}
			if err := json.Unmarshal(requestBody, &jsonBody); err == nil {
				// Add parsed JSON request body to log fields
				logFields = append(logFields, zap.Any("json", jsonBody))
				// Log request immediately using logFields
				logger.Info("RequestStart", logFields...)
			}
		}

		// Create custom ResponseWriter to capture response
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// Continue processing request
		c.Next()

		// Record response info
		latency := time.Since(start)
		// Check if Content-Type is stream data or download data
		responseContentType := c.Writer.Header().Get("Content-Type")
		if strings.Contains(responseContentType, "text/event-stream") ||
			strings.Contains(responseContentType, "application/octet-stream") ||
			strings.Contains(c.Writer.Header().Get("Content-Disposition"), "attachment") {
			// Stream data and download data only record basic info
			logger.Info("ResponseEnd",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int("status", c.Writer.Status()),
				zap.Duration("latency", latency),
			)
		} else {
			// Record complete response for other data
			fields := []zap.Field{
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int("status", c.Writer.Status()),
				zap.Duration("latency", latency),
			}

			// Only log response body if debug header is set to trace
			if strings.ToLower(c.GetHeader("debug")) == "trace" {
				fields = append(fields, zap.String("response_body", blw.body.String()))
			}

			logger.Info("ResponseEnd", fields...)
		}
	}
}

// bodyLogWriter Custom ResponseWriter to capture response body
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
