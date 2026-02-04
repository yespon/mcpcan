package model

import (
	"fmt"
	"time"
)

// SysMenu 菜单数据库模型
type SysMenu struct {
	MenuID     int64      `gorm:"primaryKey;autoIncrement;column:menu_id;comment:ID" json:"menuId"`
	PID        *int64     `gorm:"column:pid;comment:上级菜单ID" json:"pid"`
	SubCount   int        `gorm:"column:sub_count;default:0;comment:子菜单数目" json:"subCount"`
	Type       *int       `gorm:"column:type;comment:菜单类型" json:"type"`
	Title      *string    `gorm:"column:title;size:100;comment:菜单标题" json:"title"`
	Name       *string    `gorm:"column:name;size:100;comment:组件名称" json:"name"`
	Component  *string    `gorm:"column:component;size:255;comment:组件" json:"component"`
	MenuSort   *int       `gorm:"column:menu_sort;comment:排序" json:"menuSort"`
	Icon       *string    `gorm:"column:icon;size:255;comment:图标" json:"icon"`
	Path       *string    `gorm:"column:path;size:255;comment:链接地址" json:"path"`
	IFrame     *bool      `gorm:"column:i_frame;comment:是否外链" json:"iFrame"`
	Cache      *bool      `gorm:"column:cache;default:0;comment:缓存" json:"cache"`
	Hidden     *bool      `gorm:"column:hidden;default:0;comment:隐藏" json:"hidden"`
	Permission *string    `gorm:"column:permission;size:255;comment:权限" json:"permission"`
	CreateBy   *string    `gorm:"column:create_by;size:255;comment:创建者" json:"createBy"`
	UpdateBy   *string    `gorm:"column:update_by;size:255;comment:更新者" json:"updateBy"`
	CreateTime *time.Time `gorm:"column:create_time;comment:创建日期" json:"createTime"`
	UpdateTime *time.Time `gorm:"column:update_time;comment:更新时间" json:"updateTime"`
	ActivePath *string    `gorm:"column:active_path;size:255;comment:选中父级" json:"activePath"`
	EngTitle   *string    `gorm:"column:eng_title;size:255;comment:英文标题" json:"engTitle"`
}

// GetName 获取菜单名称
func (m *SysMenu) GetTitle() string {
	if m.Title == nil {
		return ""
	}
	return *m.Title
}

func (m *SysMenu) GetPermission() string {
	if m.Permission == nil {
		return ""
	}
	return *m.Permission
}

// GetType 获取菜单类型
func (m *SysMenu) GetType() string {
	if m.Type == nil {
		return ""
	}
	return fmt.Sprintf("%d", *m.Type)
}

// GetEngTitle 获取英文标题
func (m *SysMenu) GetEngTitle() string {
	if m.EngTitle == nil {
		return ""
	}
	return *m.EngTitle
}

// GetMenuSort 获取排序
func (m *SysMenu) GetMenuSort() int64 {
	if m.MenuSort == nil {
		return 0
	}
	return int64(*m.MenuSort)
}

// GetPath 获取路径
func (m *SysMenu) GetPath() string {
	if m.Path == nil {
		return ""
	}
	return *m.Path
}

// TableName 指定表名
func (SysMenu) TableName() string {
	return "sys_menu"
}
