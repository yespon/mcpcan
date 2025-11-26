package model

import (
	"encoding/json"
	"time"
)

// McpToken 定义了存储在数据库中的单个令牌记录
type McpTokens struct {
	ID               uint            `gorm:"primarykey;autoIncrement;comment:主键ID" json:"ID"`
	InstanceID       uint            `gorm:"not null;index;comment:关联的实例ID" json:"instanceID"`
	TokenType        TokenType       `gorm:"size:50;not null;comment:令牌类型" json:"tokenType"`
	Token            string          `gorm:"size:255;not null;uniqueIndex;comment:令牌的值" json:"token"`
	EnabledTransport bool            `gorm:"not null;default:false;comment:是否启用传输特定头" json:"enabledTransport"`
	Headers          json.RawMessage `gorm:"type:json;comment:自定义请求头 (JSON格式)" json:"headers,omitempty"`
	Usages           json.RawMessage `gorm:"type:json;comment:使用场景 (JSON数组格式)" json:"usages,omitempty"`
	ExpireAt         int64           `gorm:"index;comment:过期时间 (Unix时间戳)" json:"expireAt"`
	PublishAt        int64           `gorm:"comment:发布时间 (Unix时间戳)" json:"publishAt"`
	CreatedAt        time.Time       `gorm:"comment:创建时间" json:"createdAt"`
	UpdatedAt        time.Time       `gorm:"comment:更新时间" json:"updatedAt"`
}

// TableName 指定 GORM 使用的表名
func (McpTokens) TableName() string {
	return "mcpcan_tokens"
}
