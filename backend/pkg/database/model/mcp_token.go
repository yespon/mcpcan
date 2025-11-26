package model

import (
	"encoding/json"
	"time"
)

// McpToken defines a single token record stored in the database
type McpToken struct {
	ID               uint            `gorm:"primarykey;autoIncrement;comment:Primary key ID" json:"ID"`
	InstanceID       string          `gorm:"type:varchar(100);not null;default:'';comment:Associated instance ID" json:"instanceID"`
	Token            string          `gorm:"type:varchar(255);not null;default:'';comment:Value of the token" json:"token"`
	EnabledTransport bool            `gorm:"type:tinyint(1);not null;default:false;comment:Whether to enable transport-specific headers" json:"enabledTransport"`
	Headers          json.RawMessage `gorm:"type:text;comment:Custom request headers (JSON format)" json:"headers"`
	Usages           json.RawMessage `gorm:"type:text;comment:Usage scenarios (JSON array format)" json:"usages"`
	ExpireAt         int64           `gorm:"type:bigint;not null;default:0;comment:Expiration time (Millisecond timestamp)" json:"expireAt"`
	PublishAt        int64           `gorm:"type:bigint;not null;default:0;comment:Publication time (Millisecond timestamp)" json:"publishAt"`
	CreatedAt        time.Time       `gorm:"type:timestamp(3);not null;default:current_timestamp(3);comment:Creation time" json:"createdAt"`
	UpdatedAt        time.Time       `gorm:"type:timestamp(3);not null;default:current_timestamp(3) on update current_timestamp(3);comment:Update time" json:"updatedAt"`
}

// TableName 指定 GORM 使用的表名
func (McpToken) TableName() string {
	return "mcpcan_tokens"
}
