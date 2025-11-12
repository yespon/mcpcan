package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"

	"gorm.io/gorm"
)

var McpOpenapiPackageRepo *McpOpenapiPackageRepository

func init() {
	RegisterInit(func() {
		repo := NewMcpOpenapiPackageRepository()
		if err := repo.InitTable(); err != nil {
			panic(fmt.Sprintf("Failed to initialize openapi_package table: %v", err))
		}
	})
}

// McpOpenapiPackageRepository 封装 openapi_package 表的增删改查操作
type McpOpenapiPackageRepository struct{}

// NewMcpOpenapiPackageRepository 创建 McpOpenapiPackageRepository 实例
func NewMcpOpenapiPackageRepository() *McpOpenapiPackageRepository {
	McpOpenapiPackageRepo = &McpOpenapiPackageRepository{}
	return McpOpenapiPackageRepo
}

func (r *McpOpenapiPackageRepository) getDB() *gorm.DB {
	return GetDB().Model(&model.McpOpenapiPackage{})
}

// Create 创建OpenAPI文档记录
func (r *McpOpenapiPackageRepository) Create(ctx context.Context, pkg *model.McpOpenapiPackage) error {
	pkg.PrepareForCreate()
	return r.getDB().WithContext(ctx).Create(pkg).Error
}

// FindByOpenapiFileID 根据文档ID查找OpenAPI文档
func (r *McpOpenapiPackageRepository) FindByOpenapiFileID(ctx context.Context, openapiFileID string) (*model.McpOpenapiPackage, error) {
	var pkg model.McpOpenapiPackage
	if err := r.getDB().WithContext(ctx).Where("openapi_file_id = ? AND is_deleted = false", openapiFileID).First(&pkg).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("openapi package not found: %s", openapiFileID)
		}
		return nil, fmt.Errorf("failed to find openapi package: %v", err)
	}
	return &pkg, nil
}

// Update 更新OpenAPI文档记录
func (r *McpOpenapiPackageRepository) Update(ctx context.Context, pkg *model.McpOpenapiPackage) error {
	pkg.PrepareForUpdate()
	return r.getDB().WithContext(ctx).Save(pkg).Error
}

// Delete 软删除OpenAPI文档记录
func (r *McpOpenapiPackageRepository) Delete(ctx context.Context, pkg *model.McpOpenapiPackage) error {
	pkg.PrepareForDelete()
	return r.getDB().WithContext(ctx).Save(pkg).Error
}

// DeleteByOpenapiFileID 根据文档ID软删除OpenAPI文档记录
func (r *McpOpenapiPackageRepository) DeleteByOpenapiFileID(ctx context.Context, openapiFileID string) error {
	now := time.Now()
	return r.getDB().WithContext(ctx).Where("openapi_file_id = ? AND is_deleted = false", openapiFileID).
		Updates(map[string]interface{}{
			"updated_at": now,
			"is_deleted": true,
		}).Error
}

// FindAll 查找所有有效的OpenAPI文档记录
func (r *McpOpenapiPackageRepository) FindAll(ctx context.Context) ([]*model.McpOpenapiPackage, error) {
	var packages []*model.McpOpenapiPackage
	err := r.getDB().WithContext(ctx).Where("is_deleted = false").Find(&packages).Error
	if err != nil {
		return nil, err
	}
	return packages, nil
}

// FindWithPagination 分页查询OpenAPI文档记录
func (r *McpOpenapiPackageRepository) FindWithPagination(ctx context.Context, page, pageSize int32, filters map[string]interface{}) ([]*model.McpOpenapiPackage, int64, error) {
	var packages []*model.McpOpenapiPackage
	var total int64

	query := r.getDB().WithContext(ctx).Where("is_deleted = false")

	// 如果有关键词，添加搜索条件
	for key, value := range filters {
		switch key {
		case "name":
			if name, ok := value.(string); ok && name != "" {
				query = query.Where("original_name LIKE ? OR openapi_file_id LIKE ?", "%"+name+"%", "%"+name+"%")
			}
		case "types":
			if types, ok := value.([]model.OpenapiFileType); ok && len(types) > 0 {
				query = query.Where("openapi_file_type IN ?", types)
			}
		}
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&packages).Error
	if err != nil {
		return nil, 0, err
	}

	return packages, total, nil
}

// InitTable 初始化表结构
func (r *McpOpenapiPackageRepository) InitTable() error {
	// 创建表
	mod := &model.McpOpenapiPackage{}
	if err := r.getDB().AutoMigrate(mod); err != nil {
		return fmt.Errorf("failed to migrate table: %v", err)
	}

	// 检查文档ID索引是否存在
	var count int64
	sql := fmt.Sprintf("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = '%v' AND index_name = 'idx_openapi_package_file_id'", mod.TableName())
	r.getDB().Raw(sql).Count(&count)
	if count == 0 {
		// 创建文档ID索引
		sql2 := fmt.Sprintf("CREATE UNIQUE INDEX idx_openapi_package_file_id ON %v(openapi_file_id)", mod.TableName())
		if err := r.getDB().Exec(sql2).Error; err != nil {
			return fmt.Errorf("failed to create openapi_file_id index: %v", err)
		}
	}

	return nil
}
