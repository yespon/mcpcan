package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"gorm.io/gorm"
)

var SysMigrationRepo *SysMigrationRepository

func init() {
	RegisterInit(func() {
		repo := NewSysMigrationRepository(db)
		if err := repo.InitTable(); err != nil {
			panic(fmt.Sprintf("Failed to initialize sys_migrations table: %v", err))
		}
	})
}

// SysMigrationRepository migration repository
type SysMigrationRepository struct{}

// NewSysMigrationRepository creates migration repository instance
func NewSysMigrationRepository(db *gorm.DB) *SysMigrationRepository {
	SysMigrationRepo = &SysMigrationRepository{}
	return SysMigrationRepo
}

// getDB gets database connection
func (r *SysMigrationRepository) getDB() *gorm.DB {
	return GetDB().Model(&model.SysMigration{})
}

// InitTable initializes table structure
func (r *SysMigrationRepository) InitTable() error {
	return r.getDB().AutoMigrate(&model.SysMigration{})
}

// HasApplied checks if a migration has been applied
func (r *SysMigrationRepository) HasApplied(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.getDB().WithContext(ctx).Where("name = ?", name).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Apply records a migration application
func (r *SysMigrationRepository) Apply(ctx context.Context, name string) error {
	migration := &model.SysMigration{
		Name:      name,
		AppliedAt: time.Now(),
	}
	return r.getDB().WithContext(ctx).Create(migration).Error
}
