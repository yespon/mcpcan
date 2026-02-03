package model

import (
	"time"
)

// DataScope 数据权限范围类型
type DataScope string

const (
	// 全部数据权限
	DataScopeAll DataScope = "all"
	// 自定义数据权限
	DataScopeCustom DataScope = "custom"
	// 本部门数据权限
	DataScopeDept DataScope = "dept"
	// 本部门及以下数据权限
	DataScopeDeptAndChild DataScope = "dept_and_child"
	// 仅本人数据权限
	DataScopeSelf DataScope = "self"
)

// SysRole 角色数据库模型
type SysRole struct {
	RoleID      uint       `gorm:"primarykey;autoIncrement;column:role_id;comment:ID" json:"roleId"`
	Name        string     `gorm:"column:name;size:100;not null;comment:名称" json:"name"`
	Level       *int       `gorm:"column:level;comment:角色级别" json:"level"`
	Description *string    `gorm:"column:description;size:255;comment:描述" json:"description"`
	DataScope   *string    `gorm:"column:data_scope;size:255;comment:数据权限" json:"dataScope"`
	CreateBy    *string    `gorm:"column:create_by;size:255;comment:创建者" json:"createBy"`
	UpdateBy    *string    `gorm:"column:update_by;size:255;comment:更新者" json:"updateBy"`
	CreateTime  *time.Time `gorm:"column:create_time;comment:创建日期" json:"createTime"`
	UpdateTime  *time.Time `gorm:"column:update_time;comment:更新时间" json:"updateTime"`
}

// TableName 指定表名
func (SysRole) TableName() string {
	return "sys_role"
}
