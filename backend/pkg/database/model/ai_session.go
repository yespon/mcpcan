package model

import (
	"encoding/json"
	"time"
)

// AiSession 会话管理
// ModelName 在会话中指定,可以动态选择 ModelAccess 提供商下的任意模型
type AiSession struct {
	ID            int64           `gorm:"primaryKey;autoIncrement;column:id;comment:会话ID"`
	UserID        int64           `gorm:"index;column:user_id;comment:用户ID"`
	Name          string          `gorm:"size:255;column:name;comment:会话标题"`
	ModelAccessID int64           `gorm:"column:model_access_id;comment:绑定的模型配置ID"`
	ModelName     string          `gorm:"size:255;column:model_name;comment:使用的模型名称"`
	Temperature   float64         `gorm:"default:0.7;column:temperature;comment:模型温度"`
	SystemPrompt  string          `gorm:"type:text;column:system_prompt;comment:系统提示词"`
	ToolsConfig   json.RawMessage `gorm:"type:json;column:tools_config;comment:启用的工具配置"`
	MaxContext    int             `gorm:"default:20;column:max_context;comment:最大上下文条数"`
	CreateTime    time.Time       `gorm:"autoCreateTime;column:create_time;comment:创建时间"`
	UpdateTime    time.Time       `gorm:"autoUpdateTime;column:update_time;comment:更新时间"`
}

func (m *AiSession) TableName() string {
	return "mcpcan_ai_session"
}
