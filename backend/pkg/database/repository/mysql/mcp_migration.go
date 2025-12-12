package mysql

import (
	"context"
	"fmt"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"gorm.io/gorm"
)

var McpMigrationRepo *McpMigrationRepository

func init() {
	RegisterInit(func() {
		repo := NewMcpMigrationRepository()
		if err := repo.InitTable(); err != nil {
			panic(fmt.Sprintf("Failed to initialize mcpcan_tokens table: %v", err))
		}
	})
}

// NewMcpMigrationRepository creates repository and assigns global instance
func NewMcpMigrationRepository() *McpMigrationRepository {
	McpMigrationRepo = &McpMigrationRepository{}
	return McpMigrationRepo
}

type McpMigrationRepository struct{}

// getDB get db connection for migration
func (r *McpMigrationRepository) getDB() *gorm.DB {
	mod := &model.Migration{}
	return GetDB().Table(mod.TableName()).Model(mod)
}

// InitTable initializes the table schema for McpToken
func (r *McpMigrationRepository) InitTable() error {
	return r.getDB().AutoMigrate(&model.Migration{})
}

// FindByName find migration by name
func (r *McpMigrationRepository) FindByName(ctx context.Context, name string) (*model.Migration, error) {
	var migration model.Migration
	result := r.getDB().Where("name = ?", name).First(&migration)
	if result.Error != nil {
		return nil, result.Error
	}
	return &migration, nil
}

// Create create migration
func (r *McpMigrationRepository) Create(ctx context.Context, migration *model.Migration) error {
	return r.getDB().Create(migration).Error
}
