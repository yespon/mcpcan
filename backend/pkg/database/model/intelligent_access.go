package model

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
}

var tableName string

func SetIntelligentAccessTableName(name string) {
	tableName = name
}

func (m *IntelligentAccess) TableName() string {
	return tableName
}
