package model

import "time"

// SysMigration represents a database migration record
type SysMigration struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex;size:255;not null" json:"name"`
	AppliedAt time.Time `gorm:"not null" json:"appliedAt"`
}

// TableName returns the table name
func (SysMigration) TableName() string {
	return "sys_migrations"
}
