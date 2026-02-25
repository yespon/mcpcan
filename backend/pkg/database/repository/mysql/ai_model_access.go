package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"gorm.io/gorm"
)

var AiModelAccessRepo *AiModelAccessRepository

// AiModelAccessRepository 封装 ai_model_access 表的操作
type AiModelAccessRepository struct{}

// NewAiModelAccessRepository 创建 AiModelAccessRepository 实例
func NewAiModelAccessRepository() *AiModelAccessRepository {
	AiModelAccessRepo = &AiModelAccessRepository{}
	return AiModelAccessRepo
}

func (r *AiModelAccessRepository) getDB() *gorm.DB {
	return GetDB().Model(&model.AiModelAccess{})
}

// Create 创建配置
func (r *AiModelAccessRepository) Create(ctx context.Context, access *model.AiModelAccess) error {
	access.CreateTime = time.Now()
	access.UpdateTime = time.Now()
	return r.getDB().WithContext(ctx).Create(access).Error
}

// Update 更新配置
func (r *AiModelAccessRepository) Update(ctx context.Context, access *model.AiModelAccess) error {
	access.UpdateTime = time.Now()
	return r.getDB().WithContext(ctx).Where("id = ?", access.ID).Updates(access).Error
}

// Delete 删除配置
func (r *AiModelAccessRepository) Delete(ctx context.Context, id int64) error {
	return r.getDB().WithContext(ctx).Where("id = ?", id).Delete(&model.AiModelAccess{}).Error
}

// FindByID 根据ID查找
func (r *AiModelAccessRepository) FindByID(ctx context.Context, id int64) (*model.AiModelAccess, error) {
	var access model.AiModelAccess
	err := r.getDB().WithContext(ctx).First(&access, id).Error
	if err != nil {
		return nil, err
	}
	return &access, nil
}

// FindByUserID 查找用户的模型配置列表
func (r *AiModelAccessRepository) FindByUserID(ctx context.Context, userID int64) ([]*model.AiModelAccess, error) {
	var accesses []*model.AiModelAccess
	err := r.getDB().WithContext(ctx).Where("user_id = ?", userID).Find(&accesses).Error
	if err != nil {
		return nil, err
	}
	return accesses, nil
}

// InitTable 初始化表结构
func (r *AiModelAccessRepository) InitTable() error {
	if err := r.getDB().AutoMigrate(&model.AiModelAccess{}); err != nil {
		return fmt.Errorf("failed to migrate ai_model_access table: %v", err)
	}
	return nil
}
