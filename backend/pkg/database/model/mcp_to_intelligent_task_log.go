package model

import "time"

type McpToIntelligentTaskLog struct {
	ID              int64  `gorm:"primaryKey;autoIncrement;column:id;comment:ID"`
	TaskID          int64  `gorm:"not null;column:task_id;comment:任务ID"`
	McpInstanceID   string `gorm:"type:varchar(255);not null;column:mcp_instance_id;comment:MCP实例ID"`
	McpInstanceName string `gorm:"type:varchar(255);not null;column:mcp_instance_name;comment:MCP实例名称"`
	Status          bool   `gorm:"type:bool;not null;column:status;comment:任务状态"`
	ErrorLog        string `gorm:"type:text;not null;column:error_log;comment:安装错误日志"`

	IntelligentAccessID   int64  `gorm:"type:bigint;not null;column:intelligent_access_id;comment:智能体平台ID"`
	IntelligentAccessName string `gorm:"type:varchar(255);not null;column:intelligent_access_name;comment:智能体平台名称"`

	DifySpaceID   string `gorm:"type:varchar(255);not null;column:dify_space_id;comment:dify空间ID"`
	DifyUserID    string `gorm:"type:varchar(255);not null;column:dify_user_id;comment:dify用户ID"`
	DifySpaceName string `gorm:"type:varchar(255);not null;column:dify_space_name;comment:dify空间名称"`
	DifyUserName  string `gorm:"type:varchar(255);not null;column:dify_user_name;comment:dify用户名称"`

	CreatedAt time.Time `gorm:"type:timestamp(3);not null;comment:创建时间" json:"createdAt"`
	UpdatedAt time.Time `gorm:"type:timestamp(3);not null;comment:更新时间" json:"updatedAt"`
}

// TableName 指定 GORM 使用的表名
func (McpToIntelligentTaskLog) TableName() string {
	return "mcpcan_to_intelligent_task_log"
}
