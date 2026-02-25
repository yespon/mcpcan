package model

import "time"

// AiModelAccess AI模型接入配置 (API Key 配置)
// 一个配置对应一个提供商，可访问该提供商下的所有模型
type AiModelAccess struct {
	ID            int64     `gorm:"primaryKey;autoIncrement;column:id;comment:模型接入ID"`
	UserID        int64     `gorm:"index;column:user_id;comment:用户ID"`
	Name          string    `gorm:"size:255;column:name;comment:配置名称"`
	Provider      string    `gorm:"size:50;column:provider;comment:提供商: openai, qwen, zhipu, doubao, deepseek"`
	ApiKey        string    `gorm:"size:512;column:api_key;comment:API Key(加密存储)"`
	BaseUrl       string    `gorm:"size:255;column:base_url;comment:Base URL"`
	// AllowedModels 允许使用的模型ID列表（JSON数组），为空则表示不限制，可使用提供商所有模型
	AllowedModels string    `gorm:"type:text;column:allowed_models;comment:允许使用的模型ID列表(JSON数组),为空则不限制"`
	CreateTime    time.Time `gorm:"autoCreateTime;column:create_time;comment:创建时间"`
	UpdateTime    time.Time `gorm:"autoUpdateTime;column:update_time;comment:更新时间"`
}

func (m *AiModelAccess) TableName() string {
	return "mcpcan_ai_model_access"
}
