package model

import "time"

// AiModelAccess AI模型接入配置
type AiModelAccess struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;column:id;comment:模型接入ID"`
	UserID     int64     `gorm:"index;column:user_id;comment:用户ID"`
	Name       string    `gorm:"size:255;column:name;comment:配置名称"`
	Provider   string    `gorm:"size:50;column:provider;comment:提供商: openai, azure, deepseek"`
	ApiKey     string    `gorm:"size:512;column:api_key;comment:API Key(加密存储)"`
	BaseUrl    string    `gorm:"size:255;column:base_url;comment:Base URL"`
	ModelName  string    `gorm:"size:255;column:model_name;comment:模型名称"`
	CreateTime time.Time `gorm:"autoCreateTime;column:create_time;comment:创建时间"`
	UpdateTime time.Time `gorm:"autoUpdateTime;column:update_time;comment:更新时间"`
}

func (m *AiModelAccess) TableName() string {
	return "mcpcan_ai_model_access"
}
