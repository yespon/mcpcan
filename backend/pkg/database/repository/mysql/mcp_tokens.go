package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"gorm.io/gorm"
)

var McpTokenRepo *McpTokenRepository

func init() {
	RegisterInit(func() {
		repo := NewMcpTokenRepository()
		if err := repo.InitTable(); err != nil {
			panic(fmt.Sprintf("Failed to initialize mcpcan_tokens table: %v", err))
		}
	})
}

// NewMcpTokenRepository creates repository and assigns global instance
func NewMcpTokenRepository() *McpTokenRepository {
	McpTokenRepo = &McpTokenRepository{}
	return McpTokenRepo
}

type McpTokenRepository struct{}

// getDB 获取数据库连接
func (r *McpTokenRepository) getDB() *gorm.DB {
	mod := &model.McpToken{}
	return GetDB().Table(mod.TableName()).Model(mod)
}

// InitTable 初始化表结构
// InitTable migrates schema and ensures essential indexes
func (r *McpTokenRepository) InitTable() error {
	mod := &model.McpToken{}
	if err := r.getDB().AutoMigrate(mod); err != nil {
		return fmt.Errorf("failed to migrate table: %v", err)
	}

	var count int64
	table := mod.TableName()

	// Instance ID index
	sql := fmt.Sprintf("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = '%v' AND index_name = 'idx_mcp_tokens_instance_id'", table)
	r.getDB().Raw(sql).Count(&count)
	if count == 0 {
		sql2 := fmt.Sprintf("CREATE INDEX idx_mcp_tokens_instance_id ON %v(instance_id)", table)
		if err := r.getDB().Exec(sql2).Error; err != nil {
			return fmt.Errorf("failed to create idx_mcp_tokens_instance_id: %v", err)
		}
	}

	// Composite index for instance_id + expire_at
	count = 0
	sql = fmt.Sprintf("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = '%v' AND index_name = 'idx_mcp_tokens_instance_expire'", table)
	r.getDB().Raw(sql).Count(&count)
	if count == 0 {
		sql2 := fmt.Sprintf("CREATE INDEX idx_mcp_tokens_instance_expire ON %v(instance_id, expire_at)", table)
		if err := r.getDB().Exec(sql2).Error; err != nil {
			return fmt.Errorf("failed to create idx_mcp_tokens_instance_expire: %v", err)
		}
	}

	// Ensure a index for token column
	count = 0
	sql = fmt.Sprintf("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = '%v' AND index_name = 'idx_mcp_tokens_token'", table)
	r.getDB().Raw(sql).Count(&count)
	if count == 0 {
		sql2 := fmt.Sprintf("CREATE INDEX idx_mcp_tokens_token ON %v(token)", table)
		if err := r.getDB().Exec(sql2).Error; err != nil {
			return fmt.Errorf("failed to create idx_mcp_tokens_token: %v", err)
		}
	}

	return nil
}

// Create creates a single token record
func (r *McpTokenRepository) Create(ctx context.Context, t *model.McpToken) error {
	return r.getDB().WithContext(ctx).Create(t).Error
}

// CreateBatch creates multiple token records in a single call
func (r *McpTokenRepository) CreateBatch(ctx context.Context, tokens []model.McpToken) error {
	if len(tokens) == 0 {
		return nil
	}
	return r.getDB().WithContext(ctx).Create(&tokens).Error
}

// Update updates a token record by primary key
func (r *McpTokenRepository) Update(ctx context.Context, t *model.McpToken) error {
	return r.getDB().WithContext(ctx).Save(t).Error
}

// DeleteByID deletes a token by ID
func (r *McpTokenRepository) DeleteByID(ctx context.Context, id uint) error {
	return r.getDB().WithContext(ctx).Where("id = ?", id).Delete(&model.McpToken{}).Error
}

// DeleteByToken deletes a token by its value
func (r *McpTokenRepository) DeleteByToken(ctx context.Context, token string) error {
	return r.getDB().WithContext(ctx).Where("token = ?", token).Delete(&model.McpToken{}).Error
}

// DeleteByInstanceID deletes tokens belonging to an instance
func (r *McpTokenRepository) DeleteByInstanceID(ctx context.Context, instanceID uint) error {
	return r.getDB().WithContext(ctx).Where("instance_id = ?", instanceID).Delete(&model.McpToken{}).Error
}

// FindByID finds a token by ID
func (r *McpTokenRepository) FindByID(ctx context.Context, id uint) (*model.McpToken, error) {
	var t model.McpToken
	if err := r.getDB().WithContext(ctx).Where("id = ?", id).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

// FindByToken finds a token by its value
func (r *McpTokenRepository) FindByToken(ctx context.Context, token string) (*model.McpToken, error) {
	var t model.McpToken
	if err := r.getDB().WithContext(ctx).Where("token = ?", token).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

// ListByInstanceID lists tokens by instance ID
func (r *McpTokenRepository) ListByInstanceID(ctx context.Context, instanceID uint) ([]*model.McpToken, error) {
	var tokens []*model.McpToken
	if err := r.getDB().WithContext(ctx).Where("instance_id = ?", instanceID).Order("publish_at DESC, created_at DESC").Find(&tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}

// ListValidByInstanceID lists non-expired tokens by instance ID
func (r *McpTokenRepository) ListValidByInstanceID(ctx context.Context, instanceID uint) ([]*model.McpToken, error) {
	nowMs := time.Now().UnixMilli()
	var tokens []*model.McpToken
	if err := r.getDB().WithContext(ctx).
		Where("instance_id = ? AND (expire_at = 0 OR expire_at > ?)", instanceID, nowMs).
		Order("publish_at DESC, created_at DESC").
		Find(&tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}

// CountByInstanceID counts tokens by instance ID
func (r *McpTokenRepository) CountByInstanceID(ctx context.Context, instanceID uint) (int64, error) {
	var total int64
	if err := r.getDB().WithContext(ctx).Where("instance_id = ?", instanceID).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// ListWithPagination lists tokens with pagination and optional expiration filter
func (r *McpTokenRepository) ListWithPagination(ctx context.Context, instanceID uint, page, pageSize int32, includeExpired bool) ([]*model.McpToken, int64, error) {
	var tokens []*model.McpToken
	var total int64

	q := r.getDB().WithContext(ctx).Where("instance_id = ?", instanceID)
	if !includeExpired {
		q = q.Where("expire_at = 0 OR expire_at > ?", time.Now().UnixMilli())
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := q.Order("publish_at DESC, created_at DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&tokens).Error; err != nil {
		return nil, 0, err
	}

	return tokens, total, nil
}
