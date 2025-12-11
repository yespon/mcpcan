package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	golibLog "github.com/fatedier/golib/log"
	"github.com/google/uuid"
	"github.com/kymo-mcp/mcpcan/internal/gateway/proxy"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
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

		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
			c.Request.Header.Set("X-Trace-ID", traceID)
			c.Writer.Header().Set("X-Trace-ID", traceID)
		}
		logFields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
			zap.String("userAgent", c.Request.UserAgent()),
			zap.String("host", c.Request.Host),
			zap.String("traceID", traceID),
			zap.String("origin", c.GetHeader("Origin")),
			zap.String("referer", c.GetHeader("Referer")),
		}

		serversPrefix := strings.Trim(common.GetGatewayRoutePrefix(), "/")
		parts := strings.Split(strings.Trim(c.Request.URL.Path, "/"), "/")
		instanceID := ""
		if len(parts) >= 2 && parts[0] == serversPrefix {
			instanceID = parts[1]
		}
		logFields = append(logFields, zap.String("instanceID", instanceID))
		c.Request.Header.Set("X-Instance-ID", instanceID)
		c.Writer.Header().Set("X-Instance-ID", instanceID)

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

		// Log the request with all collected fields, including the masked token

		reqMsg, _ := json.Marshal(map[string]interface{}{
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"ip":          c.ClientIP(),
			"userAgent":   c.Request.UserAgent(),
			"host":        c.Request.Host,
			"origin":      c.GetHeader("Origin"),
			"referer":     c.GetHeader("Referer"),
			"headers":     headers,
			"cookies":     cookies,
			"query":       c.Request.URL.RawQuery,
			"form":        c.Request.Form,
			"pathParams":  c.Params,
			"contentType": contentType,
			"requestBody": string(requestBody),
		})

		// Extract token from headers based on priority
		var token string
		if authToken := c.GetHeader("Authorization"); authToken != "" {
			token = authToken
		} else if authToken := c.GetHeader("API-Key"); authToken != "" {
			token = authToken
		} else if authToken := c.GetHeader("X-API-Key"); authToken != "" {
			token = authToken
		}

		// Log the request with all collected fields, including the masked token
		logger.Info("Request", logFields...)

		proxy.WriteMCPLog(traceID, instanceID, token, golibLog.InfoLevel, model.EventRequest, nil, string(reqMsg))

		c.Next()

		latency := time.Since(start)

		respHeaders := make(map[string]string)
		for k, v := range c.Writer.Header() {
			if len(v) > 0 {
				respHeaders[k] = v[0]
			}
		}

		respLogFields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.Any("responseHeaders", respHeaders),
		}
		logger.Info("Response", respLogFields...)

		respMsg, _ := json.Marshal(map[string]interface{}{
			"method":          c.Request.Method,
			"path":            c.Request.URL.Path,
			"status":          c.Writer.Status(),
			"latency":         latency,
			"responseHeaders": respHeaders,
		})
		proxy.WriteMCPLog(traceID, instanceID, token, golibLog.InfoLevel, model.EventResponse, nil, string(respMsg))
	}
}

// NewServer initialize Gin engine and register all routes
func NewServer() *gin.Engine {
	r := gin.Default()

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
