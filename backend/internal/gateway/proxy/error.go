// Deprecated: 旧的网关错误处理逻辑，系统已迁移至 Traefik + Auth + Sidecar 模式。
// Deprecated: 旧的网关配置管理逻辑，系统已迁移至 Traefik + Auth + Sidecar 模式。
package proxy

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/kymo-mcp/mcpcan/pkg/logger"
	"go.uber.org/zap"
)

// Configuration switches for error handling behavior
var (
	// If true, upstream (reverse proxy) errors will be returned as MCP JSON-RPC errors.
	// If false, keep original http.Error behavior for upstream failures.
	EnableMCPErrorOnUpstreamFailure = false

	// If true, MCP error responses will use the given HTTP status code
	// instead of always responding with 200.
	UseHTTPStatusForMCPError = true

	// If true, SSE requests will output `event: error` with MCP error payload.
	EnableSSEErrorEvent = true

	// If true, include more error details in MCP error data. Otherwise sanitize.
	IncludeErrorDetails = false
)

func init() {
	EnableMCPErrorOnUpstreamFailure = envBool("MCP_ERROR_UPSTREAM_TO_JSONRPC", false)
	UseHTTPStatusForMCPError = envBool("MCP_ERROR_USE_HTTP_STATUS", true)
	EnableSSEErrorEvent = envBool("MCP_SSE_ERROR_EVENT_ENABLED", true)
	IncludeErrorDetails = envBool("MCP_ERROR_INCLUDE_DETAILS", false)
}

func envBool(key string, def bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	// support "true"/"false", "1"/"0"
	if b, err := strconv.ParseBool(v); err == nil {
		return b
	}
	vLower := strings.ToLower(v)
	return vLower == "true" || vLower == "1" || vLower == "yes"
}

// MCPError represents MCP JSON-RPC error response
type MCPError struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Error   struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data,omitempty"`
	} `json:"error"`
}

// MCPErrorObject defines the JSON-RPC error object structure used in SSE payloads.
type MCPErrorObject struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ParseMCPErrorObjectFromBytes decodes an MCP error object from raw JSON bytes.
func ParseMCPErrorObjectFromBytes(b []byte) (*MCPErrorObject, error) {
	var obj MCPErrorObject
	if err := json.Unmarshal(b, &obj); err != nil {
		return nil, err
	}
	obj.Data = SanitizeErrorData(obj.Data)
	return &obj, nil
}

// ParseMCPErrorObjectFromString decodes an MCP error object from a JSON string.
func ParseMCPErrorObjectFromString(s string) (*MCPErrorObject, error) {
	return ParseMCPErrorObjectFromBytes([]byte(s))
}

// NewMCPError creates a new MCP error response
func NewMCPError(code int, message string, data interface{}) *MCPError {
	err := &MCPError{
		JSONRPC: "2.0",
		ID:      1, // default id; can be customized if request carries JSON-RPC id
	}
	err.Error.Code = code
	err.Error.Message = message
	err.Error.Data = data
	return err
}

// MapErrorToCode maps an error to MCP JSON-RPC code
func MapErrorToCode(err error) int {
	if err == nil {
		return -32603
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "not allowed") || strings.Contains(msg, "invalid request"):
		return -32600
	case strings.Contains(msg, "not found") || strings.Contains(msg, "no such"):
		return -32601
	case strings.Contains(msg, "token") || strings.Contains(msg, "unauthorized") || strings.Contains(msg, "forbidden"):
		return -32602
	case strings.Contains(msg, "parse") || strings.Contains(msg, "json"):
		return -32700
	default:
		return -32603
	}
}

// SanitizeErrorData removes sensitive internal details from error data
func SanitizeErrorData(data interface{}) interface{} {
	if data == nil {
		return nil
	}
	if IncludeErrorDetails {
		return data
	}
	if s, ok := data.(string); ok {
		// Keep concise message; strip internals
		// Avoid leaking URLs, IPs, stack traces, SQL, etc.
		sLower := strings.ToLower(s)
		// Genericize messages
		if strings.Contains(sLower, "internal") || strings.Contains(sLower, "panic") || strings.Contains(sLower, "stack") {
			return "internal error"
		}
		// Trim overly long details
		if len(s) > 256 {
			return s[:256] + "..."
		}
		return s
	}
	// For non-string data, prefer a generic placeholder
	return "error occurred"
}

// WriteMCPError writes MCP error response to http.ResponseWriter
func WriteMCPError(w http.ResponseWriter, httpStatus int, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if UseHTTPStatusForMCPError {
		w.WriteHeader(httpStatus)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	errResp := NewMCPError(code, message, SanitizeErrorData(data))
	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		logger.Error("Failed to encode MCP error response", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// WriteMCPSSEError writes MCP error as SSE event: error
func WriteMCPSSEError(w http.ResponseWriter, code int, message string, data interface{}) {
	// Ensure SSE headers
	w.Header().Set("Content-Type", "text/event-stream;charset=UTF-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Resolve SSE error event behavior from environment
	// Event name for SSE error; default to "mcp-error" to avoid conflict with browser onerror
	eventName := strings.TrimSpace(os.Getenv("MCP_SSE_ERROR_EVENT_NAME"))
	if eventName == "" {
		eventName = "mcp-error"
	}
	// Whether to wrap payload in {"error":{...}} envelope for compatibility
	useEnvelope := envBool("MCP_SSE_ERROR_ENVELOPE", false)
	// Whether to duplicate error event: send both unnamed default event and named event
	compatDuplicate := envBool("MCP_SSE_ERROR_COMPAT_DUPLICATE", false)

	// Build error object for SSE data
	var obj *MCPErrorObject
	if s, ok := data.(string); ok {
		st := strings.TrimSpace(s)
		if strings.HasPrefix(st, "{") && strings.HasSuffix(st, "}") {
			if parsed, perr := ParseMCPErrorObjectFromString(st); perr == nil && parsed != nil {
				obj = parsed
			}
		}
	}
	if obj == nil {
		obj = &MCPErrorObject{Code: code, Message: message, Data: SanitizeErrorData(data)}
	}

	// Prepare payload JSON according to envelope setting
	var payloadBytes []byte
	if useEnvelope {
		// Envelope structure: {"error": {code, message, data}}
		envelope := struct {
			Error *MCPErrorObject `json:"error"`
		}{Error: obj}
		if b, err := json.Marshal(envelope); err == nil {
			payloadBytes = b
		} else {
			logger.Error("Failed to marshal MCP SSE error envelope", zap.Error(err))
		}
	}
	if payloadBytes == nil {
		if b, err := json.Marshal(obj); err == nil {
			payloadBytes = b
		} else {
			logger.Error("Failed to marshal MCP SSE error", zap.Error(err))
			// Fallback to JSON error object for client consistency
			fb := MCPErrorObject{Code: -32603, Message: "Internal error", Data: "internal error"}
			payloadBytes, _ = json.Marshal(fb)
		}
	}

	// Write SSE events according to compatibility settings
	if compatDuplicate {
		// Default unnamed event for clients listening via onmessage
		_, _ = w.Write([]byte("data: "))
		_, _ = w.Write(payloadBytes)
		_, _ = w.Write([]byte("\n\n"))
	}
	// Named event to avoid collision with browser onerror
	_, _ = w.Write([]byte("event: "))
	_, _ = w.Write([]byte(eventName))
	_, _ = w.Write([]byte("\n"))
	_, _ = w.Write([]byte("data: "))
	_, _ = w.Write(payloadBytes)
	_, _ = w.Write([]byte("\n\n"))

	// Flush SSE event if possible
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}
