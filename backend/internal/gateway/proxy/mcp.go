package proxy

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"
	"time"

	golibLog "github.com/fatedier/golib/log"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
	"github.com/kymo-mcp/mcpcan/pkg/redis"

	"go.uber.org/zap"
)

const (
	// DefaultReadTimeout default read timeout
	DefaultReadTimeout = 30 * time.Second
)

// McpReverseProxy multiplexed HTTP reverse proxy
type McpReverseProxy struct {
	proxy *httputil.ReverseProxy
}

// NewMCPReverseProxy create a new reverse proxy instance
func NewMCPReverseProxy() *McpReverseProxy {
	proxy := &httputil.ReverseProxy{
		Director:       director,
		ErrorHandler:   errorHandler,
		ModifyResponse: modifyResponse,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(nil),
		},
		BufferPool: newWrapPool(),
		ErrorLog:   log.New(&proxyLogger{}, "", 0),
	}

	return &McpReverseProxy{
		proxy: proxy,
	}
}

// ServeHTTP implements http.Handler interface
func (mrp *McpReverseProxy) ServeHTTP(respWriter http.ResponseWriter, req *http.Request) {
	xTraceID := req.Header.Get("X-Trace-ID")
	xInstanceID := req.Header.Get("X-Instance-ID")
	// Add panic recovery mechanism, especially for http.ErrAbortHandler
	defer func() {
		if r := recover(); r != nil {
			// Check if it is http.ErrAbortHandler
			if r == http.ErrAbortHandler {
				// This is a normal connection interruption, log at Debug level
				logger.Debug("Client connection aborted",
					zap.String("traceID", xTraceID),
					zap.String("method", req.Method),
					zap.String("path", req.URL.Path),
					zap.String("remote_addr", req.RemoteAddr),
				)
				return
			}
			// Other types of panic, log as error and re-throw
			logger.Error("Unexpected panic in ServeHTTP",
				zap.String("traceID", xTraceID),
				zap.Any("panic", r),
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path),
				zap.String("remote_addr", req.RemoteAddr),
			)
			reqAuth := &RequestAuth{}
			if v := req.Context().Value(RequestAuthKey); v != nil {
				if ra, ok := v.(*RequestAuth); ok {
					reqAuth = ra
				}
			}
			WriteMCPLog(xTraceID, xInstanceID, reqAuth.Token,
				golibLog.ErrorLevel, model.EventPanicRecovered, reqAuth.Usages,
				fmt.Sprintf("Panic recovered: %v", r))

			if strings.HasSuffix(req.URL.Path, MCP_SERVER_SUBFIX_SSE) {
				WriteMCPSSEError(respWriter, -32603, "Internal error", fmt.Sprintf("Internal Server Error: %v", r))
			} else {
				WriteMCPError(respWriter, http.StatusInternalServerError, -32603, "Internal error", fmt.Sprintf("Internal Server Error: %v", r))
			}
		}
	}()

	err := mrp.reqHandler(req)
	if err != nil {
		reqAuth := &RequestAuth{}
		if v := req.Context().Value(RequestAuthKey); v != nil {
			if ra, ok := v.(*RequestAuth); ok {
				reqAuth = ra
			}
		}
		WriteMCPLog(xTraceID, xInstanceID, reqAuth.Token,
			golibLog.ErrorLevel, model.EventRequestValidationFail, reqAuth.Usages,
			err.Error())
		// respWriter.WriteHeader(http.StatusMethodNotAllowed)
		// respWriter.Write([]byte(err.Error()))
		if strings.HasSuffix(req.URL.Path, MCP_SERVER_SUBFIX_SSE) {
			WriteMCPSSEError(respWriter, -32603, err.Error(), err.Error())
		} else {
			WriteMCPError(respWriter, http.StatusInternalServerError, -32603, err.Error(), err.Error())
		}
		return
	}

	mrp.proxy.ServeHTTP(respWriter, req)
}

func (mrp *McpReverseProxy) reqHandler(req *http.Request) error {
	xTraceID := req.Header.Get("X-Trace-ID")
	xInstanceID := req.Header.Get("X-Instance-ID")
	pathStr := req.URL.Path
	if pathStr == "" {
		return fmt.Errorf("method Not Allowed: Path is empty")
	}
	isSSEReq := false
	if strings.HasSuffix(pathStr, MCP_SERVER_SUBFIX_SSE) {
		isSSEReq = true
	}
	// Check if path prefix matches
	prefix := common.GetGatewayRoutePrefix()
	prefix = strings.Trim(prefix, "/")
	if !strings.HasPrefix(pathStr, fmt.Sprintf("/%s", prefix)) {
		return fmt.Errorf("method Not Allowed: Path Prefix is not match")
	}
	parts := strings.Split(pathStr, "/")
	instanceId := parts[2]
	// Validate if instanceId is valid
	if len(instanceId) == 0 {
		WriteMCPLog(xTraceID, xInstanceID, "", golibLog.WarnLevel, model.EventInstanceMissing, []string{},
			fmt.Sprintf("method Not Allowed: InstanceId is empty: %s", instanceId))
		return fmt.Errorf("method Not Allowed: InstanceId is empty")
	}

	// mcp config validation
	instanceInfo, err := GetInstanceInfo(instanceId)
	if err != nil {
		WriteMCPLog(xTraceID, xInstanceID, "", golibLog.WarnLevel, model.EventInstanceMissing, []string{},
			fmt.Sprintf("failed to get MCP configuration: %v for instance: %s", err.Error(), instanceId))
		return fmt.Errorf("failed to get MCP configuration: %v", err.Error())
	}
	if instanceInfo.McpConfig.Headers != nil {
		for key, value := range instanceInfo.McpConfig.Headers {
			req.Header.Set(key, value)
		}
	}

	// validate request Authorization header for instance
	reqAuth, err := validateMcpTokenForInstance(req, instanceId)
	if err != nil {
		WriteMCPLog(xTraceID, instanceId, reqAuth.Token,
			golibLog.WarnLevel, model.EventRequestValidationFail, reqAuth.Usages,
			fmt.Sprintf("failed to valid token: %v for instance: %s", err.Error(), instanceId))
		return fmt.Errorf("failed to valid token: %v", err.Error())
	}

	// Store instanceId in context
	ctx := context.WithValue(req.Context(), InstanceInfoKey, instanceInfo)
	ctx = context.WithValue(ctx, IsSSEReqKey, isSSEReq)
	ctx = context.WithValue(ctx, RequestAuthKey, reqAuth)
	*req = *req.WithContext(ctx)

	if isSSEReq {
		if instanceInfo.McpConfig.SseReadTimeout > 0 {
			ctx2, _ := context.WithTimeout(req.Context(), time.Duration(instanceInfo.McpConfig.SseReadTimeout)*time.Second)
			*req = *req.WithContext(ctx2)
		} else {
			*req = *req.WithContext(req.Context())
		}
	} else {
		timeout := instanceInfo.McpConfig.Timeout
		if timeout > 0 {
			ctx2, _ := context.WithTimeout(req.Context(), time.Duration(timeout)*time.Second)
			*req = *req.WithContext(ctx2)
		} else {
			ctx2, _ := context.WithTimeout(req.Context(), DefaultReadTimeout)
			*req = *req.WithContext(ctx2)
		}
	}
	return nil
}

type RequestAuth struct {
	TokenHeaderKey string
	Token          string
	Usages         []string
}

// director handles request modification before sending to target server
func director(req *http.Request) {
	logger.Info("Before director",
		zap.String("traceID", req.Header.Get("X-Trace-ID")),
		zap.String("method", req.Method),
		zap.String("host", req.Host),
		zap.String("url", req.URL.String()),
	)
	xTraceID := req.Header.Get("X-Trace-ID")
	xInstanceID := req.Header.Get("X-Instance-ID")
	instanceInfo, ok := req.Context().Value(InstanceInfoKey).(*InstanceInfo)
	if !ok {
		logger.Error("No InstanceInfo found in context", zap.String("traceID", xTraceID))
		WriteMCPLog(xTraceID, xInstanceID, "", golibLog.WarnLevel, model.EventInstanceMissing, []string{},
			fmt.Sprintf("no InstanceInfo in context for instance: %s", instanceInfo.InstanceID))
		return
	}
	reqAuth := &RequestAuth{}
	if v, ok := req.Context().Value(RequestAuthKey).(*RequestAuth); ok {
		reqAuth = v
	}

	isSSEReq, ok2 := req.Context().Value(IsSSEReqKey).(bool)
	if !ok2 {
		logger.Error("No IsSSEReqKey found in context", zap.String("traceID", xTraceID))
		WriteMCPLog(xTraceID, instanceInfo.InstanceID, reqAuth.Token,
			golibLog.WarnLevel, model.EventSSEFlagMissing, reqAuth.Usages,
			fmt.Sprintf("sse.flag.missing for instance: %s", instanceInfo.InstanceID))
		return
	}
	parts := strings.Split(req.URL.Path, "/")
	pathNum := len(parts)
	if pathNum <= 2 {
		logger.Error("Path is too short", zap.String("traceID", xTraceID))
		WriteMCPLog(xTraceID, instanceInfo.InstanceID, reqAuth.Token,
			golibLog.WarnLevel, model.EventPathTooShort, reqAuth.Usages,
			fmt.Sprintf("path.too.short for instance: %s", instanceInfo.InstanceID))
		return
	}

	prefix := getProxyPrefix(instanceInfo.InstanceID)
	targetUrl, err := url.Parse(instanceInfo.McpConfig.URL)
	if err != nil {
		logger.Error("Failed to parse URL", zap.String("traceID", xTraceID), zap.Error(err))
		WriteMCPLog(xTraceID, instanceInfo.InstanceID, reqAuth.Token,
			golibLog.ErrorLevel, model.EventUpstreamURLParseFail, reqAuth.Usages,
			fmt.Sprintf("Failed to parse URL: %s for instance: %s", err.Error(), instanceInfo.InstanceID))
		return
	}

	switch instanceInfo.AccessType {
	case model.AccessTypeHosting:
		switch instanceInfo.ProxyMcpProtocol {
		case model.McpProtocolSSE:
			if isSSEReq {
				handleHostingSSEReq(req, instanceInfo, targetUrl)
			} else {
				handleHostingSSEReqForEvent(req, instanceInfo, prefix, targetUrl)
			}
		case model.McpProtocolStreamableHttp:
			handleHostingStreamableHTTPReq(req, instanceInfo, targetUrl)
		default:
			logger.Error("McpProtocol is not supported", zap.String("traceID", xTraceID))
			WriteMCPLog(xTraceID, instanceInfo.InstanceID, reqAuth.Token,
				golibLog.WarnLevel, model.EventProtocolUnsupported, reqAuth.Usages,
				fmt.Sprintf("McpProtocol is not supported for instance: %s", instanceInfo.InstanceID))
			return
		}
	case model.AccessTypeProxy:
		switch instanceInfo.ProxyMcpProtocol {
		case model.McpProtocolSSE:
			if isSSEReq {
				handleProxySSEReq(req, instanceInfo, targetUrl)
			} else {
				handleProxySSEReqForEvent(req, prefix, targetUrl)
			}
		case model.McpProtocolStreamableHttp:
			handleProxyStreamableHTTPPathReq(req, instanceInfo, targetUrl)
		default:
			logger.Error("McpProtocol is not supported", zap.String("traceID", xTraceID))
			WriteMCPLog(xTraceID, instanceInfo.InstanceID, reqAuth.Token,
				golibLog.WarnLevel, model.EventProtocolUnsupported, reqAuth.Usages,
				fmt.Sprintf("McpProtocol is not supported for instance: %s", instanceInfo.InstanceID))
			return
		}
	default:
		logger.Error("AccessType is not supported", zap.String("traceID", xTraceID))
		WriteMCPLog(xTraceID, instanceInfo.InstanceID, reqAuth.Token,
			golibLog.WarnLevel, model.EventAccessUnsupported, reqAuth.Usages,
			fmt.Sprintf("AccessType is not supported for instance: %s", instanceInfo.InstanceID))
		return
	}
	// Log request info
	logger.Info("After director",
		zap.String("traceID", xTraceID),
		zap.String("instance_id", instanceInfo.InstanceID),
		zap.Bool("is_ssereq", isSSEReq),
		zap.String("url", req.URL.String()),
	)
}

// Handle response modification before sending to client
func modifyResponse(resp *http.Response) error {
	xTraceID := resp.Request.Header.Get("X-Trace-ID")
	// Check if it is SSE response
	if strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream") {
		// Get instanceId from context
		instanceInfo, ok := resp.Request.Context().Value(InstanceInfoKey).(*InstanceInfo)
		if !ok {
			return &proxyError{
				message: "instanceId not found in context",
				status:  http.StatusInternalServerError,
			}
		}
		reqAuth := &RequestAuth{}
		if v := resp.Request.Context().Value(RequestAuthKey); v != nil {
			if ra, ok := v.(*RequestAuth); ok {
				reqAuth = ra
			}
		}
		WriteMCPLog(xTraceID, instanceInfo.InstanceID, reqAuth.Token,
			golibLog.InfoLevel, model.EventSSEStart, reqAuth.Usages,
			fmt.Sprintf("sse.start for instance: %s", instanceInfo.InstanceID))

		// Check if request context has been canceled
		select {
		case <-resp.Request.Context().Done():
			ctxErr := resp.Request.Context().Err()
			logger.Debug("Request context canceled in modifyResponse", zap.String("traceID", xTraceID), zap.Error(ctxErr))
			WriteMCPLog(xTraceID, instanceInfo.InstanceID, reqAuth.Token,
				golibLog.InfoLevel, model.EventSSECancel, reqAuth.Usages,
				fmt.Sprintf("sse.cancel for instance: %s", instanceInfo.InstanceID))
			return &proxyError{
				message: "request context canceled",
				status:  http.StatusRequestTimeout,
			}
		default:
		}

		// "text/event-stream;charset=UTF-8"
		// Set necessary SSE response headers
		resp.Header.Set("Content-Type", "text/event-stream;charset=UTF-8")
		resp.Header.Set("Cache-Control", "no-cache")
		resp.Header.Set("Connection", "keep-alive")
		resp.Header.Set("X-Accel-Buffering", "no")
		resp.Header.Set("Transfer-Encoding", "chunked")

		var reader io.Reader = resp.Body
		var err error

		// Check if it is Gzip compressed
		if resp.Header.Get("Content-Encoding") == "gzip" {
			// Use gzip.Reader to wrap original response body, it will auto-decompress
			reader, err = gzip.NewReader(resp.Body)
			if err != nil {
				WriteMCPLog(xTraceID, instanceInfo.InstanceID, reqAuth.Token,
					golibLog.ErrorLevel, model.EventGzipReaderFailed, reqAuth.Usages,
					fmt.Sprintf("failed to create gzip reader: %v", err))
				return &proxyError{
					message: fmt.Sprintf("failed to create gzip reader: %v", err),
					status:  http.StatusInternalServerError,
				}
			}
			// We have handled decompression, so remove it from response header to prevent downstream (e.g. browser) from decompressing again
			resp.Header.Del("Content-Encoding")
		}

		host := resp.Request.Host

		// Replace response body with our custom Reader
		resp.Body = io.NopCloser(&SSEResponseBodyReader{
			host:    host,
			src:     reader,
			info:    instanceInfo,
			reqAuth: reqAuth,
			resp:    resp,
		})

		// Ensure response header allows chunked transfer
		resp.Header.Del("Content-Length")
	}

	return nil
}

// SSEResponseBodyReader wraps original response body, adds instanceID before each SSE message
type SSEResponseBodyReader struct {
	host    string
	src     io.Reader     // Decompressed original response body
	buffer  bytes.Buffer  // Used for buffering data and processing
	reader  *bufio.Reader // Convenient for reading by line or delimiter
	reqAuth *RequestAuth
	info    *InstanceInfo
	resp    *http.Response
}

func (r *SSEResponseBodyReader) Read(p []byte) (n int, err error) {
	// Initialize reader on first Read call
	if r.reader == nil {
		r.reader = bufio.NewReader(r.src)
	}
	xTraceID := r.resp.Request.Header.Get("X-Trace-ID")

	// Continuously read data from source until p is filled or error occurs
	for {
		// Read data from our buffer (if any)
		if r.buffer.Len() > 0 {
			return r.buffer.Read(p)
		}

		// Buffer is empty, read next SSE message from source
		// SSE messages are separated by `\n\n`
		msgBytes, readErr := r.reader.ReadBytes('\n')
		if readErr != nil && readErr != io.EOF {
			// Check if it is connection-related error
			if isConnectionError(readErr) {
				// Connection interruption is normal, return EOF
				logger.Debug("SSE connection interrupted", zap.String("traceID", xTraceID), zap.Error(readErr))
				return 0, io.EOF
			}
			// Other read errors
			return 0, readErr
		}

		// Continue reading until message boundary `\n\n` is encountered
		for {
			line, err := r.reader.ReadBytes('\n')
			msgBytes = append(msgBytes, line...)
			if err != nil {
				// Check if it is connection-related error
				if isConnectionError(err) {
					logger.Debug("SSE connection interrupted during message read", zap.String("traceID", xTraceID), zap.Error(err))
					readErr = io.EOF
				} else {
					readErr = err
				}
				break
			}
			if len(bytes.TrimSpace(line)) == 0 { // Message ends
				break
			}
		}

		if len(msgBytes) > 0 {
			msgStr := string(msgBytes)
			// Handle SSE messages of type event: endpoint
			if strings.Contains(msgStr, "event: endpoint") || strings.Contains(msgStr, "event:pathParams") {
				// Add prefix proxy rule
				// If contains data: / , replace with data: /{prefix}/
				// If contains data:/ , replace with data: /{prefix}/
				prefix := getProxyPrefix(r.info.InstanceID)
				if strings.Contains(msgStr, "data: /") {
					msgBytes = bytes.ReplaceAll(msgBytes, []byte("data: /"), []byte(fmt.Sprintf("data: /%s/", strings.Trim(prefix, "/"))))
				} else if strings.Contains(msgStr, "data:/") {
					msgBytes = bytes.ReplaceAll(msgBytes, []byte("data:/"), []byte(fmt.Sprintf("data:/%s/", strings.Trim(prefix, "/"))))
				} else if strings.Contains(msgStr, "data: ?") {
					msgBytes = bytes.ReplaceAll(msgBytes, []byte("data: ?"), []byte(fmt.Sprintf("data: /%s?", strings.Trim(prefix, "/"))))
				} else if strings.Contains(msgStr, "data:?") {
					msgBytes = bytes.ReplaceAll(msgBytes, []byte("data:?"), []byte(fmt.Sprintf("data:/%s?", strings.Trim(prefix, "/"))))
				}
				logger.Info("Replace SSE event:endpoint", zap.String("traceID", xTraceID), zap.String("old", msgStr), zap.String("new", string(msgBytes)))
				WriteMCPLog(xTraceID, r.info.InstanceID, r.reqAuth.Token,
					golibLog.DebugLevel, model.EventSSEEndpointRewrite, r.reqAuth.Usages,
					fmt.Sprintf("endpoint rewritten for instance: %s", r.info.InstanceID))
			}

			// Write modified data into internal buffer
			r.buffer.Write(msgBytes)
		}

		// If source is exhausted (EOF) and our buffer is also empty, return EOF
		if readErr == io.EOF && r.buffer.Len() == 0 {
			WriteMCPLog(xTraceID, r.info.InstanceID, r.reqAuth.Token,
				golibLog.DebugLevel, model.EventSSEEof, r.reqAuth.Usages,
				fmt.Sprintf("sse eof for instance: %s", r.info.InstanceID))
			return 0, io.EOF
		}

		// If other error occurs, return directly
		if readErr != nil && readErr != io.EOF {
			WriteMCPLog(xTraceID, r.info.InstanceID, r.reqAuth.Token,
				golibLog.WarnLevel, model.EventSSEReadError, r.reqAuth.Usages,
				fmt.Sprintf("sse read error for instance: %s, error: %v", r.info.InstanceID, readErr))
			return 0, readErr
		}
	}
}

// isConnectionError checks if error is related to connection interruption
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	errorStr := err.Error()
	return strings.Contains(errorStr, "connection reset") ||
		strings.Contains(errorStr, "broken pipe") ||
		strings.Contains(errorStr, "connection refused") ||
		strings.Contains(errorStr, "context canceled") ||
		strings.Contains(errorStr, "context deadline exceeded") ||
		strings.Contains(errorStr, "use of closed network connection")
}

// Get proxy prefix
func getProxyPrefix(instanceID string) string {
	prefix := common.GetGatewayRoutePrefix()
	prefix = path.Join(prefix, instanceID)
	return prefix
}

// Hosting mode, SSE long connection request handling
func handleHostingSSEReq(req *http.Request, instanceInfo *InstanceInfo, targetUrl *url.URL) string {
	req.URL.Scheme = targetUrl.Scheme
	req.URL.Host = targetUrl.Host
	req.URL.Path = targetUrl.Path
	// Append RawQuery
	if targetUrl.RawQuery != "" {
		req.URL.RawQuery = req.URL.RawQuery + "&" + targetUrl.RawQuery
	}
	// Append header
	if instanceInfo.McpConfig.Headers != nil {
		for key, value := range instanceInfo.McpConfig.Headers {
			req.Header.Set(key, value)
		}
	}
	return req.URL.Path
}

// Hosting mode, SSE event request handling
func handleHostingSSEReqForEvent(req *http.Request, instanceInfo *InstanceInfo, prefix string, targetUrl *url.URL) string {
	req.URL.Scheme = targetUrl.Scheme
	req.URL.Host = targetUrl.Host
	if strings.HasPrefix(req.URL.Path, path.Join(prefix)) {
		req.URL.Path = strings.Replace(req.URL.Path, path.Join(prefix), "", 1)
	}

	if instanceInfo.AccessType == model.AccessTypeHosting &&
		instanceInfo.McpProtocol == model.McpProtocolStdio &&
		strings.Contains(instanceInfo.Instance.ImgAddr, common.DefatuleHostingImg) {
		req.URL.Path = strings.TrimRight(req.URL.Path, "/") + "/"
	}
	return req.URL.Path
}

// Hosting mode, Streamable HTTP request handling
func handleHostingStreamableHTTPReq(req *http.Request, instanceInfo *InstanceInfo, targetUrl *url.URL) string {
	req.URL.Scheme = targetUrl.Scheme
	req.URL.Host = targetUrl.Host
	req.URL.Path = targetUrl.Path
	// Append RawQuery
	if targetUrl.RawQuery != "" {
		req.URL.RawQuery = req.URL.RawQuery + "&" + targetUrl.RawQuery
	}
	if instanceInfo.AccessType == model.AccessTypeHosting &&
		instanceInfo.McpProtocol == model.McpProtocolStdio &&
		strings.Contains(instanceInfo.Instance.ImgAddr, common.DefatuleHostingImg) {
		req.URL.Path = strings.TrimRight(req.URL.Path, "/") + "/"
	}
	return req.URL.Path
}

// Proxy mode, SSE long connection request handling
func handleProxySSEReq(req *http.Request, instanceInfo *InstanceInfo, targetUrl *url.URL) string {
	req.URL.Scheme = targetUrl.Scheme
	req.URL.Host = targetUrl.Host
	req.URL.Path = targetUrl.Path
	// Append RawQuery
	if targetUrl.RawQuery != "" {
		req.URL.RawQuery = req.URL.RawQuery + "&" + targetUrl.RawQuery
	}
	// Append header
	if instanceInfo.McpConfig.Headers != nil {
		for key, value := range instanceInfo.McpConfig.Headers {
			req.Header.Set(key, value)
		}
	}
	return req.URL.Path
}

// Proxy mode, SSE event request handling
func handleProxySSEReqForEvent(req *http.Request, prefix string, targetUrl *url.URL) string {
	req.URL.Scheme = targetUrl.Scheme
	req.URL.Host = targetUrl.Host
	if strings.HasPrefix(req.URL.Path, path.Join(prefix)) {
		req.URL.Path = strings.Replace(req.URL.Path, path.Join(prefix), "", 1)
	}
	return req.URL.Path
}

// Proxy mode, Streamable HTTP request handling
func handleProxyStreamableHTTPPathReq(req *http.Request, instanceInfo *InstanceInfo, targetUrl *url.URL) string {
	req.URL.Scheme = targetUrl.Scheme
	req.URL.Host = targetUrl.Host
	req.URL.Path = targetUrl.Path
	// Append RawQuery
	if targetUrl.RawQuery != "" {
		req.URL.RawQuery = req.URL.RawQuery + "&" + targetUrl.RawQuery
	}
	// Append header
	if instanceInfo.McpConfig.Headers != nil {
		for key, value := range instanceInfo.McpConfig.Headers {
			req.Header.Set(key, value)
		}
	}
	return req.URL.Path
}

func GetInstanceInfo(instanceID string) (*InstanceInfo, error) {
	// Use business layer cache service to get instance info
	service := NewMcpInstanceService()
	return service.GetInstanceInfo(instanceID)
}

func extractJSONRPCParams(req *http.Request) string {
	if req.Body == nil {
		return ""
	}
	b, err := io.ReadAll(req.Body)
	if err != nil {
		return ""
	}
	req.Body = io.NopCloser(bytes.NewBuffer(b))
	s := strings.TrimSpace(string(b))
	if s == "" {
		return ""
	}
	type jr struct {
		Params any `json:"params"`
	}
	var j jr
	if err := json.Unmarshal(b, &j); err != nil {
		return ""
	}
	pb, _ := json.Marshal(j.Params)
	return string(pb)
}

// validateMcpTokenForInstance 校验请求 Authorization header 是否有效
func validateMcpTokenForInstance(req *http.Request, instanceID string) (*RequestAuth, error) {
	// delete Authorization header from req
	ra := &RequestAuth{}

	var tokenHeaderKey string
	var token string
	if req.Header.Get("Authorization") != "" {
		tokenHeaderKey = "Authorization"
		token = req.Header.Get("Authorization")
	} else if req.Header.Get("API-Key") != "" {
		tokenHeaderKey = "API-Key"
		token = req.Header.Get("API-Key")
	} else if req.Header.Get("X-API-Key") != "" {
		tokenHeaderKey = "X-API-Key"
		token = req.Header.Get("X-API-Key")
	}
	req.Header.Del(tokenHeaderKey)

	if token == "" {
		return ra, fmt.Errorf("instance %v enabled token but request %v header is empty", instanceID, tokenHeaderKey)
	}

	tokenCache := redis.GetMcpTokenCache()
	cacheKey := tokenCache.GenerateCacheKey(instanceID, token)
	var mcpToken *model.McpToken
	if v := tokenCache.GetRedis(cacheKey); v != nil {
		mcpToken = v
	} else {
		trow, err := mysql.McpTokenRepo.FindByToken(context.Background(), instanceID, token)
		if err != nil || trow == nil || trow.InstanceID != instanceID {
			return ra, fmt.Errorf("failed to get MCP token: %v for instance %v", "not found", instanceID)
		}
		_ = tokenCache.SetRedis(cacheKey, trow, redis.TokenCacheExpire)
		mcpToken = trow
	}

	if mcpToken.EnabledTransport {
		headers := map[string]string{}
		if len(mcpToken.Headers) > 0 {
			_ = json.Unmarshal(mcpToken.Headers, &headers)
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	var usages []string
	if len(mcpToken.Usages) > 0 {
		_ = json.Unmarshal(mcpToken.Usages, &usages)
	}
	ra = &RequestAuth{
		TokenHeaderKey: tokenHeaderKey,
		Token:          token,
		Usages:         usages,
	}
	return ra, nil
}
