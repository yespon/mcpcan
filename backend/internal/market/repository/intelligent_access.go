package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"gorm.io/gorm"
)

var IntelligentAccessRepo *IntelligentAccessRepository

func init() {
	mysql.RegisterInit(func() {
		repo := NewIntelligentAccessRepository()
		if err := repo.InitTable(); err != nil {
			panic(fmt.Sprintf("Failed to initialize intelligent_access table: %v", err))
		}
	})
}

// IntelligentAccessRepository 封装 intelligent_access 表的增删改查操作
type IntelligentAccessRepository struct{}

// NewIntelligentAccessRepository 创建 IntelligentAccessRepository 实例
func NewIntelligentAccessRepository() *IntelligentAccessRepository {
	IntelligentAccessRepo = &IntelligentAccessRepository{}
	return IntelligentAccessRepo
}

// getDB 获取数据库连接
func (r *IntelligentAccessRepository) getDB() *gorm.DB {
	return mysql.GetDB().Model(&model.IntelligentAccess{})
}

// Create 创建
func (r *IntelligentAccessRepository) Create(ctx context.Context, intelligentAccess *model.IntelligentAccess) error {
	intelligentAccess.CreateTime = time.Now()
	intelligentAccess.UpdateTime = time.Now()
	return r.getDB().WithContext(ctx).Create(intelligentAccess).Error
}

// Update 更新
func (r *IntelligentAccessRepository) Update(ctx context.Context, intelligentAccess *model.IntelligentAccess) error {
	intelligentAccess.UpdateTime = time.Now()
	return r.getDB().WithContext(ctx).Where("access_id = ?", intelligentAccess.ID).Updates(intelligentAccess).Error
}

// Delete 删除
func (r *IntelligentAccessRepository) Delete(ctx context.Context, accessId int64) error {
	return r.getDB().WithContext(ctx).Where("access_id = ?", accessId).Delete(&model.IntelligentAccess{}).Error
}

// FindByID 根据ID查找
func (r *IntelligentAccessRepository) FindByID(ctx context.Context, accessId int64) (*model.IntelligentAccess, error) {
	var intelligentAccess model.IntelligentAccess
	err := r.getDB().WithContext(ctx).First(&intelligentAccess, accessId).Error
	if err != nil {
		return nil, err
	}
	return &intelligentAccess, nil
}

// FindAll 查找所有
func (r *IntelligentAccessRepository) FindAll(ctx context.Context) ([]*model.IntelligentAccess, error) {
	var intelligentAccesss []*model.IntelligentAccess
	err := r.getDB().WithContext(ctx).Find(&intelligentAccesss).Error
	if err != nil {
		return nil, err
	}
	return intelligentAccesss, nil
}

// InitTable 初始化表结构
func (r *IntelligentAccessRepository) InitTable() error {
	// 创建表
	mod := &model.IntelligentAccess{}
	if err := r.getDB().AutoMigrate(mod); err != nil {
		return fmt.Errorf("failed to migrate table: %v", err)
	}
	return nil
}
