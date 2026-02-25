package model

import (
	"time"
)

// DeptSource 部门来源类型
type DeptSource string

const (
	// 自建部门
	DeptSourcePlatform DeptSource = "PLATFORM"
	// 飞书部门
	DeptSourceFeishu DeptSource = "FEISHU"
)

// SysDept 部门数据库模型
type SysDept struct {
	DeptID     uint       `gorm:"primarykey;autoIncrement;column:dept_id;comment:ID" json:"deptId"`
	PID        *uint      `gorm:"column:pid;comment:上级部门" json:"pid"`
	SubCount   int        `gorm:"column:sub_count;default:0;comment:子部门数目" json:"subCount"`
	Name       string     `gorm:"column:name;size:255;not null;comment:名称" json:"name"`
	DeptSort   int        `gorm:"column:dept_sort;default:999;comment:排序" json:"deptSort"`
	Enabled    int        `gorm:"column:enabled;not null;comment:状态" json:"enabled"`
	CreateBy   *string    `gorm:"column:create_by;size:255;comment:创建者" json:"createBy"`
	UpdateBy   *string    `gorm:"column:update_by;size:255;comment:更新者" json:"updateBy"`
	CreateTime *time.Time `gorm:"column:create_time;comment:创建日期" json:"createTime"`
	UpdateTime *time.Time `gorm:"column:update_time;comment:更新时间" json:"updateTime"`
	ImageURL   *string    `gorm:"column:image_url;size:255;comment:图片" json:"imageUrl"`
	Source     DeptSource `gorm:"column:source;size:32;not null;comment:部门来源 PLATFORM：自建，FEISHU:飞书" json:"source"`
}

// TableName 指定表名
func (SysDept) TableName() string {
	return "sys_dept"
}
