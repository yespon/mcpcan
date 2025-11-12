package proxy

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/fatedier/golib/log"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
)

func writeMCPLog(instanceID string, tokenType model.TokenType, token string, level log.Level, event string, message string, extra map[string]any) {
	payload := map[string]any{
		"ts":      time.Now().Format(time.RFC3339Nano),
		"event":   event,
		"message": message,
	}
	logBytes, _ := json.Marshal(payload)
	var extraBytes json.RawMessage
	if extra != nil {
		if b, err := json.Marshal(extra); err == nil {
			extraBytes = b
		}
	}
	_ = RecordGatewayLog(instanceID, tokenType, strings.TrimSpace(token), level, logBytes, extraBytes)
}

func writeMCPLogWithUsage(instanceID string, tokenHeaderKey string, token string, usages []string, level log.Level, event string, message string, extra map[string]any) {
	payload := map[string]any{
		"ts":      time.Now().Format(time.RFC3339Nano),
		"event":   event,
		"message": message,
	}
	logBytes, _ := json.Marshal(payload)
	if extra == nil {
		extra = map[string]any{}
	}
	extra["tokenHeaderKey"] = strings.TrimSpace(tokenHeaderKey)
	var extraBytes json.RawMessage
	if b, err := json.Marshal(extra); err == nil {
		extraBytes = b
	}
	_ = RecordGatewayLogWithUsage(instanceID, model.TokenType(""), strings.TrimSpace(token), usages, level, logBytes, extraBytes)
}
