package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"gorm.io/gorm"
)

var SysRoleRepo *SysRoleRepository

// SysRoleRepository 封装 sys_role 表的增删改查操作
type SysRoleRepository struct{}

// NewSysRoleRepository 创建 SysRoleRepository 实例
func NewSysRoleRepository() *SysRoleRepository {
	SysRoleRepo = &SysRoleRepository{}
	return SysRoleRepo
}

// getDB 获取数据库连接
func (r *SysRoleRepository) getDB() *gorm.DB {
	return GetDB().Model(&model.SysRole{})
}

// Create 创建角色
func (r *SysRoleRepository) Create(ctx context.Context, role *model.SysRole) error {
	now := time.Now()
	role.CreateTime = &now
	role.UpdateTime = &now
	return r.getDB().WithContext(ctx).Create(role).Error
}

// Update 更新角色
func (r *SysRoleRepository) Update(ctx context.Context, role *model.SysRole) error {
	now := time.Now()
	role.UpdateTime = &now
	return r.getDB().WithContext(ctx).Where("role_id = ?", role.RoleID).Save(role).Error
}

// Delete 删除角色
func (r *SysRoleRepository) Delete(ctx context.Context, roleID uint) error {
	return r.getDB().WithContext(ctx).Where("role_id = ?", roleID).Delete(&model.SysRole{}).Error
}

// BatchDelete 批量删除角色
func (r *SysRoleRepository) BatchDelete(ctx context.Context, roleIDs []uint) error {
	return r.getDB().WithContext(ctx).Where("role_id IN ?", roleIDs).Delete(&model.SysRole{}).Error
}

func (r *SysRoleRepository) FindByIDs(ctx context.Context, roleIDs []uint) ([]*model.SysRole, error) {
	var roles []*model.SysRole
	err := r.getDB().WithContext(ctx).Where("role_id IN ?", roleIDs).Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// FindByID 根据ID查找角色
func (r *SysRoleRepository) FindByID(ctx context.Context, roleID uint) (*model.SysRole, error) {
	var role model.SysRole
	err := r.getDB().WithContext(ctx).Where("role_id = ?", roleID).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// FindByName 根据名称查找角色
func (r *SysRoleRepository) FindByName(ctx context.Context, name string) (*model.SysRole, error) {
	var role model.SysRole
	err := r.getDB().WithContext(ctx).Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// FindWithPagination 分页查询角色（支持角色名称和描述模糊查询）
func (r *SysRoleRepository) FindWithPagination(ctx context.Context, page, pageSize int, keyword string, ids []uint) ([]*model.SysRole, int64, error) {
	var roles []*model.SysRole
	var total int64

	query := r.getDB().WithContext(ctx)

	if keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if len(ids) > 0 {
		query = query.Where("role_id IN (?)", ids)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("role_id ASC").Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// InitTable 初始化表结构
func (r *SysRoleRepository) InitTable() error {
	mod := &model.SysRole{}
	if err := r.getDB().AutoMigrate(mod); err != nil {
		return fmt.Errorf("failed to migrate table %s: %v", mod.TableName(), err)
	}
	return nil
}
