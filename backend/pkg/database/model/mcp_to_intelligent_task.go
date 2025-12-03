package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type HeaderInfo struct {
	Token   string            `gorm:"type:varchar(255);not null;column:token;comment:token"`
	Headers map[string]string `gorm:"type:json;not null;column:headers;comment:headers"`
}

type InsertIntelligentInfo struct {
	DifySpaceID   string `gorm:"type:varchar(255);not null;column:dify_space_id;comment:dify空间ID"`
	DifyUserID    string `gorm:"type:varchar(255);not null;column:dify_user_id;comment:dify用户ID"`
	DifySpaceName string `gorm:"type:varchar(255);not null;column:dify_space_name;comment:dify空间名称"`
	DifyUserName  string `gorm:"type:varchar(255);not null;column:dify_user_name;comment:dify用户名称"`

	Headers map[string]*HeaderInfo `gorm:"type:json;not null;column:headers;comment:headers"`
}

type InsertIntelligentInfos []*InsertIntelligentInfo

type InstallLog struct {
	McpInstanceID         string                  `gorm:"type:varchar(255);not null;column:mcp_instance_id;comment:MCP实例ID"`
	McpInstanceName       string                  `gorm:"type:varchar(255);not null;column:mcp_instance_id;comment:MCP实例名称"`
	InsertIntelligentLogs []*InsertIntelligentLog `gorm:"type:json;not null;column:insert_intelligent_info;comment:智能体平台插入日志"`
	Status                bool                    `gorm:"type:bool;not null;column:status;comment:任务状态"`
	ErrorLog              string                  `gorm:"type:text;not null;column:error_log;comment:安装错误日志"`
}

type InsertIntelligentLog struct {
	InsertIntelligentInfo *InsertIntelligentInfo `gorm:"type:json;not null;column:insert_intelligent_info;comment:智能体平台插入信息"`
	ErrorLog              string                 `gorm:"type:text;not null;column:error_log;comment:安装错误日志"`
	Status                bool                   `gorm:"type:bool;not null;column:status;comment:任务状态"`
}

type InstallLogs []*InstallLog

type StringSlice []string

// McpToIntelligentTask mcp到智能体平台任务
type McpToIntelligentTask struct {
	ID                     int64                  `gorm:"primaryKey;autoIncrement;column:id;comment:接入信息ID"`
	Desc                   string                 `gorm:"type:varchar(255);not null;column:desc;comment:任务描述"`
	IntelligentAccessID    int64                  `gorm:"type:bigint;not null;column:intelligent_access_id;comment:智能体平台ID"`
	IntelligentAccessName  string                 `gorm:"type:varchar(255);not null;column:intelligent_access_name;comment:智能体平台名称"`
	InsertIntelligentInfos InsertIntelligentInfos `gorm:"type:json;not null;column:insert_intelligent_info;comment:智能体平台插入信息"`
	McpInstanceIDs         StringSlice            `gorm:"type:json;not null;column:mcp_instance_ids;comment:MCP实例ID列表"`
	Status                 string                 `gorm:"type:varchar(255);not null;column:status;comment:任务状态"`
	InstallLogs            InstallLogs            `gorm:"type:json;not null;column:install_logs;comment:安装日志"`
	Domain                 string                 `gorm:"type:varchar(255);not null;column:domain;comment:mcp访问域名"`
	CreatedAt              time.Time              `gorm:"type:timestamp(3);not null;comment:创建时间" json:"createdAt"`
	UpdatedAt              time.Time              `gorm:"type:timestamp(3);not null;comment:更新时间" json:"updatedAt"`
}

func (s *StringSlice) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &s)
}
func (s StringSlice) Value() (driver.Value, error) {
	bytes, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// InsertIntelligentInfos 实现 GORM 接口
func (i *InsertIntelligentInfos) Scan(value interface{}) error {
	if value == nil {
		*i = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, i)
}

func (i InsertIntelligentInfos) Value() (driver.Value, error) {
	if i == nil {
		return nil, nil
	}
	return json.Marshal(i)
}

// InstallLogs 实现 GORM 接口
func (i *InstallLogs) Scan(value interface{}) error {
	if value == nil {
		*i = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, i)
}

func (i InstallLogs) Value() (driver.Value, error) {
	if i == nil {
		return nil, nil
	}
	return json.Marshal(i)
}

// InstallLog 实现 GORM 接口
func (i *InstallLog) Scan(value interface{}) error {
	if value == nil {
		*i = InstallLog{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, i)
}

func (i InstallLog) Value() (driver.Value, error) {
	return json.Marshal(i)
}

// InsertIntelligentLog 实现 GORM 接口
func (i *InsertIntelligentLog) Scan(value interface{}) error {
	if value == nil {
		*i = InsertIntelligentLog{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, i)
}

func (i InsertIntelligentLog) Value() (driver.Value, error) {
	return json.Marshal(i)
}

// InsertIntelligentInfo 实现 GORM 接口
func (i *InsertIntelligentInfo) Scan(value interface{}) error {
	if value == nil {
		*i = InsertIntelligentInfo{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, i)
}

func (i InsertIntelligentInfo) Value() (driver.Value, error) {
	return json.Marshal(i)
}

func (m *McpToIntelligentTask) TableName() string {
	return "mcpcan_to_intelligent_task"
}
