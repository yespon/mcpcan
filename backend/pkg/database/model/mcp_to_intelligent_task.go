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
	SpaceID   string `gorm:"type:varchar(255);not null;column:space_id;comment:空间ID"`
	UserID    string `gorm:"type:varchar(255);not null;column:user_id;comment:用户ID"`
	SpaceName string `gorm:"type:varchar(255);not null;column:space_name;comment:空间名称"`
	UserName  string `gorm:"type:varchar(255);not null;column:user_name;comment:用户名称"`

	Headers map[string]*HeaderInfo `gorm:"type:json;not null;column:headers;comment:headers"`
}

type InsertIntelligentInfos []*InsertIntelligentInfo

type StringSlice []string

// McpToIntelligentTask mcp到智能体平台任务
type McpToIntelligentTask struct {
	ID                     int64                  `gorm:"primaryKey;autoIncrement;column:id;comment:接入信息ID"`
	Creator                int64                  `gorm:"type:bigint;not null;column:creator;comment:创建者ID"`
	Desc                   string                 `gorm:"type:varchar(255);not null;column:desc;comment:任务描述"`
	IntelligentAccessID    int64                  `gorm:"type:bigint;not null;column:intelligent_access_id;comment:智能体平台ID"`
	IntelligentAccessName  string                 `gorm:"type:varchar(255);not null;column:intelligent_access_name;comment:智能体平台名称"`
	InsertIntelligentInfos InsertIntelligentInfos `gorm:"type:json;not null;column:insert_intelligent_info;comment:智能体平台插入信息"`
	McpInstanceIDs         StringSlice            `gorm:"type:json;not null;column:mcp_instance_ids;comment:MCP实例ID列表"`
	Status                 string                 `gorm:"type:varchar(255);not null;column:status;comment:任务状态"`
	Domain                 string                 `gorm:"type:varchar(255);not null;column:domain;comment:mcp访问域名"`
	Cookie                 string                 `gorm:"type:text;not null;column:cookie;comment:cookie"`
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
