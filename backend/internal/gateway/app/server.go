package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kymo-mcp/mcpcan/internal/gateway/proxy"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestResponseLoggingMiddleware detailed request-response logging middleware
func RequestResponseLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// record request start time
		start := time.Now()

		// read request body
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// reset request body so subsequent handlers can read it
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}
		// prepare log fields
		logFields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		// record request headers
		headers := make(map[string]string)
		for k, v := range c.Request.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}
		logFields = append(logFields, zap.Any("headers", headers))

		// record query parameters
		if c.Request.URL.RawQuery != "" {
			logFields = append(logFields, zap.String("query", c.Request.URL.RawQuery))
		}

		// record form parameters
		if err := c.Request.ParseForm(); err == nil {
			if len(c.Request.Form) > 0 {
				logFields = append(logFields, zap.Any("form", c.Request.Form))
			}
		}

		// check if Content-Type is JSON and attempt to parse request body
		contentType := c.GetHeader("Content-Type")
		if strings.Contains(contentType, "application/json") && len(requestBody) > 0 {
			var jsonBody interface{}
			if err := json.Unmarshal(requestBody, &jsonBody); err == nil {
				logFields = append(logFields, zap.Any("json", jsonBody))
			}
		}

		// create custom ResponseWriter to capture response
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// continue processing request
		c.Next()

		// record response info
		latency := time.Since(start)
		// check if Content-Type is streaming or download data
		responseContentType := c.Writer.Header().Get("Content-Type")
		if strings.Contains(responseContentType, "text/event-stream") ||
			strings.Contains(responseContentType, "application/octet-stream") ||
			strings.Contains(c.Writer.Header().Get("Content-Disposition"), "attachment") {
			// for streaming and download data, log only basic info
			logger.Info("request completed",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int("status", c.Writer.Status()),
				zap.Duration("latency", latency),
			)
		} else {
			// for other data, log full response
			logger.Info("request completed",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int("status", c.Writer.Status()),
				zap.Duration("latency", latency),
				zap.String("response_body", blw.body.String()),
			)
		}
	}
}

// bodyLogWriter custom ResponseWriter to capture response body
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// CORSMiddleware allows all origins, methods, and headers
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// NewServer initialize Gin engine and register all routes
func NewServer() *gin.Engine {
	r := gin.Default()

	// add CORS middleware to allow all origins and methods
	r.Use(CORSMiddleware())
	// add request-response logging middleware
	r.Use(RequestResponseLoggingMiddleware())

	// get route prefix
	serversPrefix := common.GetGatewayRoutePrefix()
	serversPrefix = strings.Trim(serversPrefix, "/")

	// register MCP service SSE protocol reverse proxy
	mcpSSEServerProxy := proxy.NewMCPReverseProxy()
	r.Any(fmt.Sprintf("/%s/*path", serversPrefix), gin.WrapH(mcpSSEServerProxy))

	// health check
	r.GET("/health", func(c *gin.Context) { c.String(200, "ok") })

	return r
}
