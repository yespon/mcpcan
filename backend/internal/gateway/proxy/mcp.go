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
	"net/textproto"
	"net/url"
	"path"
	"strings"
	"time"

	golibLog "github.com/fatedier/golib/log"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/logger"

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
	parts0 := strings.Split(req.URL.Path, "/")
	var iid0 string
	if len(parts0) > 2 {
		iid0 = parts0[2]
	}
	// Add panic recovery mechanism, especially for http.ErrAbortHandler
	defer func() {
		if r := recover(); r != nil {
			// Check if it is http.ErrAbortHandler
			if r == http.ErrAbortHandler {
				// This is a normal connection interruption, log at Debug level
				logger.Debug("Client connection aborted",
					zap.String("method", req.Method),
					zap.String("path", req.URL.Path),
					zap.String("remote_addr", req.RemoteAddr),
				)
				return
			}
			// Other types of panic, log as error and re-throw
			logger.Error("Unexpected panic in ServeHTTP",
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
			writeMCPLog(iid0, reqAuth.TokenHeaderKey, reqAuth.Token,
				golibLog.ErrorLevel, model.EventPanicRecovered, reqAuth.Usages,

				buildLogFromReq(req, "panic"))
			// panic(r) // Re-throw panic that is not ErrAbortHandler
			// respWriter.WriteHeader(http.StatusInternalServerError)
			// respWriter.Write([]byte(fmt.Sprintf("Internal Server Error: %v", r)))
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
		writeMCPLog(iid0, reqAuth.TokenHeaderKey, reqAuth.Token,
			golibLog.ErrorLevel, model.EventRequestValidationFail, reqAuth.Usages,

			buildLogFromReq(req, err.Error()))
		// respWriter.WriteHeader(http.StatusMethodNotAllowed)
		// respWriter.Write([]byte(err.Error()))
		if strings.HasSuffix(req.URL.Path, MCP_SERVER_SUBFIX_SSE) {
			WriteMCPSSEError(respWriter, -32603, err.Error(), err.Error())
		} else {
			WriteMCPError(respWriter, http.StatusInternalServerError, -32603, err.Error(), err.Error())
		}
		return
	}

	if info := req.Context().Value(InstanceInfoKey); info != nil {
		reqAuth := &RequestAuth{}
		if v := req.Context().Value(RequestAuthKey); v != nil {
			if ra, ok := v.(*RequestAuth); ok {
				reqAuth = ra
			}
		}
		writeMCPLog(iid0, reqAuth.TokenHeaderKey, reqAuth.Token,
			golibLog.InfoLevel, model.EventRequestReceived, reqAuth.Usages,

			buildLogFromReq(req, "received"))
	}

	mrp.proxy.ServeHTTP(respWriter, req)
}

func (mrp *McpReverseProxy) reqHandler(req *http.Request) error {
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
		writeMCPLog(instanceId, "", "", golibLog.WarnLevel, model.EventInstanceMissing, []string{},
			buildLogFromReq(req, "InstanceId is empty"))
		return fmt.Errorf("method Not Allowed: InstanceId is empty")
	}

	// mcp config validation
	instanceInfo, err := GetInstanceInfo(instanceId)
	if err != nil {
		writeMCPLog(instanceId, "", "", golibLog.WarnLevel, model.EventInstanceMissing, []string{},
			buildLogFromReq(req, err.Error()))
		return fmt.Errorf("failed to get MCP configuration: %v", err.Error())
	}

	// validate request Authorization header for instance
	reqAuth, err := validReqAuthorizationForInstance(req, instanceInfo)
	if err != nil {
		writeMCPLog(instanceId, reqAuth.TokenHeaderKey, reqAuth.Token,
			golibLog.WarnLevel, model.EventRequestValidationFail, reqAuth.Usages,

			buildLogFromReq(req, err.Error()))
		return fmt.Errorf("failed to valid token: %v", err.Error())
	}

	if instanceInfo.McpConfig.Headers != nil {
		for key, value := range instanceInfo.McpConfig.Headers {
			req.Header.Set(key, value)
		}
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
		zap.String("method", req.Method),
		zap.String("host", req.Host),
		zap.String("url", req.URL.String()),
	)
	instanceInfo, ok := req.Context().Value(InstanceInfoKey).(*InstanceInfo)
	if !ok {
		logger.Error("No InstanceInfo found in context")
		writeMCPLog(instanceInfo.InstanceID, "", "", golibLog.WarnLevel, model.EventInstanceMissing, []string{},
			buildLogFromReq(req, "no InstanceInfo in context"))
		return
	}
	reqAuth := &RequestAuth{}
	if v, ok := req.Context().Value(RequestAuthKey).(*RequestAuth); ok {
		reqAuth = v
	}
	writeMCPLog(instanceInfo.InstanceID, reqAuth.TokenHeaderKey, reqAuth.Token,
		golibLog.InfoLevel, model.EventDirectorBefore, reqAuth.Usages,

		buildLogFromReq(req, "director.before"))

	isSSEReq, ok2 := req.Context().Value(IsSSEReqKey).(bool)
	if !ok2 {
		logger.Error("No IsSSEReqKey found in context")
		writeMCPLog(instanceInfo.InstanceID, reqAuth.TokenHeaderKey, reqAuth.Token,
			golibLog.WarnLevel, model.EventSSEFlagMissing, reqAuth.Usages,

			buildLogFromReq(req, "sse.flag.missing"))
		return
	}

	parts := strings.Split(req.URL.Path, "/")
	pathNum := len(parts)
	if pathNum <= 2 {
		logger.Error("Path is too short")
		writeMCPLog(instanceInfo.InstanceID, reqAuth.TokenHeaderKey, reqAuth.Token,
			golibLog.WarnLevel, model.EventPathTooShort, reqAuth.Usages,

			buildLogFromReq(req, "path.too.short"))
		return
	}

	prefix := getProxyPrefix(instanceInfo.InstanceID)

	targetUrl, err := url.Parse(instanceInfo.McpConfig.URL)
	if err != nil {
		logger.Error("Failed to parse URL", zap.Error(err))
		writeMCPLog(instanceInfo.InstanceID, reqAuth.TokenHeaderKey, reqAuth.Token,
			golibLog.ErrorLevel, model.EventUpstreamURLParseFail, reqAuth.Usages,

			buildLogFromReq(req, "Failed to parse URL: "+err.Error()))
		return
	}

	switch instanceInfo.AccessType {
	case model.AccessTypeHosting:
		switch instanceInfo.McpProtocol {
		case model.McpProtocolSSE:
			if isSSEReq {
				handleHostingSSEReq(req, instanceInfo, targetUrl)
			} else {
				handleHostingSSEReqForEvent(req, instanceInfo, prefix, targetUrl)
			}
		case model.McpProtocolStreamableHttp:
			handleHostingStreamableHTTPReq(req, instanceInfo, targetUrl)
		default:
			logger.Error("McpProtocol is not supported")
			writeMCPLog(instanceInfo.InstanceID, reqAuth.TokenHeaderKey, reqAuth.Token,
				golibLog.WarnLevel, model.EventProtocolUnsupported, reqAuth.Usages,

				buildLogFromReq(req, "McpProtocol is not supported"))
			return
		}
	case model.AccessTypeProxy:
		switch instanceInfo.McpProtocol {
		case model.McpProtocolSSE:
			if isSSEReq {
				handleProxySSEReq(req, instanceInfo, targetUrl)
			} else {
				handleProxySSEReqForEvent(req, prefix, targetUrl)
			}
		case model.McpProtocolStreamableHttp:
			handleProxyStreamableHTTPPathReq(req, instanceInfo, targetUrl)
		default:
			logger.Error("McpProtocol is not supported")
			writeMCPLog(instanceInfo.InstanceID, reqAuth.TokenHeaderKey, reqAuth.Token,
				golibLog.WarnLevel, model.EventProtocolUnsupported, reqAuth.Usages,

				buildLogFromReq(req, "McpProtocol is not supported"))
			return
		}
	default:
		logger.Error("AccessType is not supported")
		writeMCPLog(instanceInfo.InstanceID, reqAuth.TokenHeaderKey, reqAuth.Token,
			golibLog.WarnLevel, model.EventAccessUnsupported, reqAuth.Usages,

			buildLogFromReq(req, "AccessType is not supported"))
		return
	}
	// Log request info
	logger.Info("After director",
		zap.String("instance_id", instanceInfo.InstanceID),
		zap.Bool("is_ssereq", isSSEReq),
		zap.String("url", req.URL.String()),
	)

	writeMCPLog(instanceInfo.InstanceID, reqAuth.TokenHeaderKey, reqAuth.Token,
		golibLog.InfoLevel, model.EventDirectorAfter, reqAuth.Usages,
		buildLogFromReq(req, "director.after"))
}

// Handle response modification before sending to client
func modifyResponse(resp *http.Response) error {
	// Check if it is SSE response
	isSSEReq, ok := resp.Request.Context().Value(IsSSEReqKey).(bool)
	if ok && isSSEReq {
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
		writeMCPLog(instanceInfo.InstanceID, reqAuth.TokenHeaderKey, reqAuth.Token,
			golibLog.InfoLevel, model.EventSSEStart, reqAuth.Usages,
			buildLogFromReq(resp.Request, "sse start"))

		// Check if request context has been canceled
		select {
		case <-resp.Request.Context().Done():
			ctxErr := resp.Request.Context().Err()
			logger.Debug("Request context canceled in modifyResponse", zap.Error(ctxErr))
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
				writeMCPLog(instanceInfo.InstanceID, reqAuth.TokenHeaderKey, reqAuth.Token,
					golibLog.ErrorLevel, model.EventGzipReaderFailed, reqAuth.Usages,
					buildLogFromReq(resp.Request, err.Error()))
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
				logger.Debug("SSE connection interrupted", zap.Error(readErr))
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
					logger.Debug("SSE connection interrupted during message read", zap.Error(err))
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
				logger.Info("Replace SSE event:endpoint", zap.String("old", msgStr), zap.String("new", string(msgBytes)))
				writeMCPLog(r.info.InstanceID, r.reqAuth.TokenHeaderKey, r.reqAuth.Token,
					golibLog.DebugLevel, model.EventSSEEndpointRewrite, r.reqAuth.Usages,
					buildLogFromReq(r.resp.Request, "endpoint rewritten"))
			}

			// Write modified data into internal buffer
			r.buffer.Write(msgBytes)
		}

		// If source is exhausted (EOF) and our buffer is also empty, return EOF
		if readErr == io.EOF && r.buffer.Len() == 0 {
			writeMCPLog(r.info.InstanceID, r.reqAuth.TokenHeaderKey, r.reqAuth.Token,
				golibLog.DebugLevel, model.EventSSEEof, r.reqAuth.Usages,
				buildLogFromReq(r.resp.Request, "sse eof"))
			return 0, io.EOF
		}

		// If other error occurs, return directly
		if readErr != nil && readErr != io.EOF {
			writeMCPLog(r.info.InstanceID, r.reqAuth.TokenHeaderKey, r.reqAuth.Token,
				golibLog.WarnLevel, model.EventSSEReadError, r.reqAuth.Usages,
				buildLogFromReq(r.resp.Request, readErr.Error()))
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
	// Append header
	if instanceInfo.McpConfig.Headers != nil {
		for key, value := range instanceInfo.McpConfig.Headers {
			req.Header.Set(key, value)
		}
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

func buildLogFromReq(req *http.Request, msg string) *model.Log {
	isSSE := strings.HasSuffix(req.URL.Path, MCP_SERVER_SUBFIX_SSE)
	params := extractJSONRPCParams(req)
	return &model.Log{
		Message: msg,
		URL:     req.URL.String(),
		Method:  req.Method,
		Path:    req.URL.Path,
		Params:  params,
		IsSSE:   isSSE,
		TS:      time.Now().Format(time.RFC3339Nano),
	}
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

// validReqAuthorizationForInstance 校验请求 Authorization header 是否有效
func validReqAuthorizationForInstance(req *http.Request, instanceInfo *InstanceInfo) (*RequestAuth, error) {
	// delete Authorization header from req
	ra := &RequestAuth{}

	// 1. instance 是否启用了 token 认证，没有则直接返回成功
	if !instanceInfo.EnabledToken {
		return ra, nil
	}

	// 2. instance 启用了 token 认证，校验 token 是否有效
	if len(instanceInfo.Tokens) == 0 {
		return ra, fmt.Errorf("instance %v enabled token but token list is empty", instanceInfo.Instance.InstanceID)
	}

	tokenHeaderKey := ""
	token := ""
	for _, tokenInfo := range instanceInfo.Tokens {
		tokenHeaderKey = tokenInfo.ToTokenHeaderKey()
		canonKey := textproto.CanonicalMIMEHeaderKey(tokenHeaderKey)
		token = strings.TrimSpace(req.Header.Get(canonKey))
		if tokenInfo.Token == token {
			// 3. token 校验通过，校验 token 是否过期
			expireAt := time.Unix(0, tokenInfo.ExpireAt*int64(time.Millisecond))
			if tokenInfo.ExpireAt > 0 && time.Now().After(expireAt) {
				return ra, fmt.Errorf("instance %v enabled token validate but token expired", instanceInfo.Instance.InstanceID)
			}
			if tokenInfo.EnabledTransport {
				for k, v := range tokenInfo.Headers {
					if k != tokenHeaderKey {
						req.Header.Set(k, v)
					}
				}
			} else {
				req.Header.Del(canonKey)
			}
			ra = &RequestAuth{
				TokenHeaderKey: tokenHeaderKey,
				Token:          token,
				Usages:         tokenInfo.Usages,
			}
			return ra, nil
		}
	}
	if len(token) == 0 {
		return ra, fmt.Errorf("instance %v enabled token but request %v header is empty", instanceInfo.Instance.InstanceID, tokenHeaderKey)
	}

	return ra, fmt.Errorf("instance %v enabled token validate but token not found", instanceInfo.Instance.InstanceID)
}
