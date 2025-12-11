package model

import "time"

// Migration records the status of a database migration task
type Migration struct {
	ID          uint      `gorm:"primarykey"`
	Name        string    `gorm:"size:255;not null;uniqueIndex;comment:unique name of the migration task"`
	CompletedAt time.Time `gorm:"comment:time when the migration was completed"`
}

// TableName specifies the table name used by GORM
func (Migration) TableName() string {
	return "mcpcan_migrations"
}
