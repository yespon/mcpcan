package model

import (
	"encoding/json"
	"time"

	"github.com/fatedier/golib/log"
)

type TokenType string

func (tokenType TokenType) String() string {
	return string(tokenType)
}

const (
	// API token type
	TokenTypeBearer TokenType = "bearer"
	// token type
	TokenTypeBasic TokenType = "basic"
	// token type
	TokenTypeKey TokenType = "api-key"
	// x-api-key
	TokenTypeXAPIKey TokenType = "x-api-key"
)

type GatewayLog struct {
	ID         uint            `gorm:"primaryKey"`
	InstanceID string          `gorm:"size:100;not null;comment:instance ID" json:"instanceID"`
	TokenType  TokenType       `gorm:"size:100;not null;default:'';comment:token type" json:"tokenType"`
	Token      string          `gorm:"size:1000;not null;default:'';comment:token" json:"token"`
	Usages     string          `gorm:"size:1000;not null;default:'';comment:usage scenarios" json:"usages"`
	Extra      json.RawMessage `gorm:"type:json;not null;comment:extra information" json:"extra"`
	Log        string          `gorm:"type:text;not null;comment:log details" json:"log"`
	Level      log.Level       `gorm:"type:int;not null;default:0;comment:log level" json:"level"`
	CreatedAt  time.Time       `gorm:"type:timestamp(3);not null;comment:creation time" json:"createdAt"`
	UpdatedAt  time.Time       `gorm:"type:timestamp(3);not null;comment:update time" json:"updatedAt"`
}

func (gatewayLog *GatewayLog) TableName() string {
	return "mcp_gateway_log"
}
