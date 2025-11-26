package model

import "time"

// Migration 记录数据库迁移任务的状态
type Migration struct {
	ID          uint      `gorm:"primarykey"`
	Name        string    `gorm:"size:255;not null;uniqueIndex;comment:迁移任务的唯一名称"`
	CompletedAt time.Time `gorm:"comment:迁移完成时间"`
}

// TableName 指定 GORM 使用的表名
func (Migration) TableName() string {
	return "mcpcan_migrations"
}
