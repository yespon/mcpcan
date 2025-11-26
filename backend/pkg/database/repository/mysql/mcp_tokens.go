package mysql

import (
	"fmt"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"gorm.io/gorm"
)

var McpTokensRepo *McpTokensRepository

func init() {
	RegisterInit(func() {
		repo := NewMcpTokensRepository()
		if err := repo.InitTable(); err != nil {
			panic(fmt.Sprintf("Failed to initialize mcp_template table: %v", err))
		}
	})
}

func NewMcpTokensRepository() *McpTokensRepository {
	return &McpTokensRepository{}
}

type McpTokensRepository struct{}

// getDB 获取数据库连接
func (r *McpTokensRepository) getDB() *gorm.DB {
	mod := &model.McpTokens{}
	return GetDB().Table(mod.TableName()).Model(mod)
}

// InitTable 初始化表结构
func (r *McpTokensRepository) InitTable() error {
	// 创建表
	mod := &model.McpTokens{}
	if err := r.getDB().AutoMigrate(mod); err != nil {
		return fmt.Errorf("failed to migrate table: %v", err)
	}

	// 检查索引是否存在
	var count int64
	sql := fmt.Sprintf("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = '%v' AND index_name = 'idx_mcp_tokens_instance_id'", (&model.McpTokens{}).TableName())
	r.getDB().Raw(sql).Count(&count)
	if count == 0 {
		// 创建索引
		sql2 := fmt.Sprintf("CREATE INDEX idx_mcp_tokens_instance_id ON %v(instance_id)", (&model.McpTokens{}).TableName())
		if err := r.getDB().Exec(sql2).Error; err != nil {
			return fmt.Errorf("failed to create index: %v", err)
		}
	}

	return nil
}
