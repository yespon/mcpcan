package model

import (
	"encoding/json"
	"time"

	"github.com/fatedier/golib/log"
)

type Event string

const (
	EventRequest                 Event = "request"
	EventResponse                Event = "response"
	EventPanicRecovered          Event = "panic.recovered"
	EventRequestValidationFail   Event = "request.validation.failed"
	EventDirectorBefore          Event = "director.before"
	EventDirectorAfter           Event = "director.after"
	EventInstanceMissing         Event = "instance.missing"
	EventSSEFlagMissing          Event = "sse.flag.missing"
	EventPathTooShort            Event = "path.too.short"
	EventUpstreamURLParseFail    Event = "upstream.url.parse.failed"
	EventProtocolUnsupported     Event = "protocol.unsupported"
	EventAccessUnsupported       Event = "access.unsupported"
	EventSSEStart                Event = "sse.start"
	EventSSECancel               Event = "sse.cancel"
	EventGzipReaderFailed        Event = "gzip.reader.failed"
	EventSSEEndpointRewrite      Event = "sse.endpoint.rewrite"
	EventSSEEof                  Event = "sse.eof"
	EventSSEReadError            Event = "sse.read.error"
	EventClientCanceled          Event = "client.canceled"
	EventProxyErrorLog           Event = "proxy.error.log"
	EventUpstreamConnInterrupted Event = "upstream.connection.interrupted"
	EventUpstreamError           Event = "upstream.error"
	EventAuthSuccess             Event = "auth.success"
	EventAuthFailed              Event = "auth.failed"
)

type GatewayLog struct {
	ID         uint            `gorm:"primaryKey"`
	TraceID    string          `gorm:"size:100;not null;default:'';comment:trace ID" json:"traceID"`
	InstanceID string          `gorm:"size:100;not null;comment:instance ID" json:"instanceID"`
	ToolName   string          `gorm:"size:100;not null;default:'';comment:tool name" json:"toolName"`
	Token      string          `gorm:"size:1000;not null;default:'';comment:token" json:"token"`
	Usages     string          `gorm:"size:1000;not null;default:'';comment:usage scenarios" json:"usages"`
	Log        json.RawMessage `gorm:"type:text;not null;comment:log details" json:"log"`
	Level      log.Level       `gorm:"type:int;not null;default:0;comment:log level" json:"level"`
	Event      Event           `gorm:"size:100;not null;default:'';comment:event type" json:"event"`
	CreatedAt  time.Time       `gorm:"type:timestamp(3);not null;comment:creation time" json:"createdAt"`
	UpdatedAt  time.Time       `gorm:"type:timestamp(3);not null;comment:update time" json:"updatedAt"`
}

type Log struct {
	Event   Event     `json:"event"`
	Level   log.Level `json:"level"`
	Message string    `json:"message"`
	TS      string    `json:"ts"`
}

func (gatewayLog *GatewayLog) TableName() string {
	return "mcpcan_gateway_log"
}
