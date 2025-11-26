package model

import (
	"encoding/json"
	"time"
)

// McpToken defines a single token record stored in the database
type McpTokens struct {
	ID               uint            `gorm:"primarykey;autoIncrement;comment:Primary key ID" json:"ID"`
	InstanceID       uint            `gorm:"type:varchar(100);not null;default:'';comment:Associated instance ID" json:"instanceID"`
	TokenKey         string          `gorm:"type:varchar(255);not null;default:'';comment:Unique key for the token" json:"tokenKey"`
	Token            string          `gorm:"type:varchar(255);not null;default:'';comment:Value of the token" json:"token"`
	EnabledTransport bool            `gorm:"type:tinyint(1);not null;default:false;comment:Whether to enable transport-specific headers" json:"enabledTransport"`
	Headers          json.RawMessage `gorm:"type:text;not null;default:'';comment:Custom request headers (JSON format)" json:"headers,omitempty"`
	Usages           json.RawMessage `gorm:"type:text;not null;default:'';comment:Usage scenarios (JSON array format)" json:"usages,omitempty"`
	ExpireAt         int64           `gorm:"type:bigint;not null;default:0;comment:Expiration time (Millisecond timestamp)" json:"expireAt"`
	PublishAt        int64           `gorm:"type:bigint;not null;default:0;comment:Publication time (Millisecond timestamp)" json:"publishAt"`
	CreatedAt        time.Time       `gorm:"type:timestamp(3);not null;default:current_timestamp(3);comment:Creation time" json:"createdAt"`
	UpdatedAt        time.Time       `gorm:"type:timestamp(3);not null;default:current_timestamp(3) on update current_timestamp(3);comment:Update time" json:"updatedAt"`
}

// TableName 指定 GORM 使用的表名
func (McpTokens) TableName() string {
	return "mcpcan_tokens"
}

type McpToken struct {
	ID               uint              `json:"id"`
	Token            string            `json:"token"`
	Headers          map[string]string `json:"headers,omitempty"`
	EnabledTransport bool              `json:"enabledTransport"`
	ExpireAt         int64             `json:"expireAt"`
	PublishAt        int64             `json:"publishAt"`
	Usages           []string          `json:"usages"`
}
