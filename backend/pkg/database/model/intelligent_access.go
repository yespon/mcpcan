package model

import "time"

// IntelligentAccess 接入信息
type IntelligentAccess struct {
	ID         int64  `gorm:"primaryKey;autoIncrement;column:access_id;comment:接入信息ID"`
	AccessName string `gorm:"type:varchar(255);not null;column:access_name;comment:名称"`
	AccessType string `gorm:"type:varchar(255);not null;column:access_type;comment:类型: COZE, DifyEnterprise"`
	DbHost     string `gorm:"type:varchar(255);column:db_host;comment:dify数据库地址"`
	DbPort     int    `gorm:"type:int;column:db_port;comment:dify数据库端口"`
	DbUser     string `gorm:"type:varchar(32);column:db_user;comment:dify数据库帐号"`
	DbPassword string `gorm:"type:varchar(255);column:db_password;comment:dify数据库密码"`
	DbName     string `gorm:"type:varchar(255);column:db_name;comment:dify数据库名"`

	CreateTime time.Time `gorm:"type:datetime;column:create_time;comment:创建时间"`
	UpdateTime time.Time `gorm:"type:datetime;column:update_time;comment:更新时间"`
}

var tableName string

func SetIntelligentAccessTableName(name string) {
	tableName = name
}

func (m *IntelligentAccess) TableName() string {
	return tableName
}
