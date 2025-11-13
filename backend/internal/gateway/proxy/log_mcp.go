package proxy

import (
    "encoding/json"
    "strings"
    "time"

    "github.com/fatedier/golib/log"
    "github.com/kymo-mcp/mcpcan/pkg/database/model"
)

func maskToken(s string) string {
    ss := strings.TrimSpace(s)
    n := len(ss)
    if n <= 8 {
        if n == 0 {
            return ""
        }
        return "****"
    }
    return ss[:4] + "****" + ss[n-4:]
}

func writeMCPLog(instanceID string, tokenType model.TokenType, token string, level log.Level, event string, message string, ctx map[string]any, extra map[string]any) {
    payload := map[string]any{
        "ts":     time.Now().Format(time.RFC3339Nano),
        "event":  event,
        "message": message,
    }
    for k, v := range ctx {
        payload[k] = v
    }
    logBytes, _ := json.Marshal(payload)
    var extraBytes json.RawMessage
    if extra != nil {
        if b, err := json.Marshal(extra); err == nil {
            extraBytes = b
        }
    }
    _ = RecordGatewayLog(instanceID, tokenType, maskToken(token), level, logBytes, extraBytes)
}

