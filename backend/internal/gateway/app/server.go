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
		start := time.Now()

		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		logFields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
			zap.String("userAgent", c.Request.UserAgent()),
			zap.String("host", c.Request.Host),
			zap.String("origin", c.GetHeader("Origin")),
			zap.String("referer", c.GetHeader("Referer")),
		}

		headers := make(map[string]string)
		for k, v := range c.Request.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}
		logFields = append(logFields, zap.Any("headers", headers))

		cookies := make(map[string]string)
		for _, ck := range c.Request.Cookies() {
			cookies[ck.Name] = ck.Value
		}
		if len(cookies) > 0 {
			logFields = append(logFields, zap.Any("cookies", cookies))
		}

		if c.Request.URL.RawQuery != "" {
			logFields = append(logFields, zap.String("query", c.Request.URL.RawQuery))
		}

		if err := c.Request.ParseForm(); err == nil {
			if len(c.Request.Form) > 0 {
				logFields = append(logFields, zap.Any("form", c.Request.Form))
			}
		}

		if len(c.Params) > 0 {
			pathParams := make(map[string]string)
			for _, p := range c.Params {
				pathParams[p.Key] = p.Value
			}
			logFields = append(logFields, zap.Any("pathParams", pathParams))
		}

		contentType := c.GetHeader("Content-Type")
		if strings.Contains(contentType, "application/json") && len(requestBody) > 0 {
			var jsonBody interface{}
			if err := json.Unmarshal(requestBody, &jsonBody); err == nil {
				logFields = append(logFields, zap.Any("json", jsonBody))
			}
		}

		logger.Info("Request", logFields...)

		c.Next()

		latency := time.Since(start)

		respHeaders := make(map[string]string)
		for k, v := range c.Writer.Header() {
			if len(v) > 0 {
				respHeaders[k] = v[0]
			}
		}
		logger.Info("Response",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.Any("responseHeaders", respHeaders),
		)
	}
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
