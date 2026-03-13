package model

import "time"

// SysDataPermission 数据权限配置模型
type SysDataPermission struct {
	ID          int64      `gorm:"primaryKey;autoIncrement;column:id;comment:ID" json:"id"`
	DataType    string     `gorm:"column:data_type;size:50;not null;index:idx_query,priority:1;comment:数据类型" json:"dataType"`
	DataID      string     `gorm:"column:data_id;size:64;not null;index:idx_query,priority:2;comment:具体的数据记录ID" json:"dataId"`
	TargetType  string     `gorm:"column:target_type;size:20;not null;index:idx_query,priority:3;comment:授权目标类型(ALL, USER, DEPT, ROLE)" json:"targetType"`
	TargetID    int64      `gorm:"column:target_id;not null;index:idx_query,priority:4;comment:授权目标ID" json:"targetId"`
	IsBlacklist int        `gorm:"column:is_blacklist;type:tinyint(1);not null;default:0;comment:是否为黑名单(0:否, 1:是)" json:"isBlacklist"`
	CreatedBy   int64      `gorm:"column:created_by;comment:创建人ID" json:"createdBy"`
	CreatedAt   *time.Time `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"createdAt"`
}

// TableName 指定表名
func (SysDataPermission) TableName() string {
	return "sys_data_permission"
}
