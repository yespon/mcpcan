package mysql

import (
	"context"
	"fmt"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"gorm.io/gorm"
)

var SysRolesMenusRepo *SysRolesMenusRepository

// SysRolesMenusRepository 角色菜单关联仓库
type SysRolesMenusRepository struct{}

// NewSysRolesMenusRepository 创建角色菜单关联仓库实例
func NewSysRolesMenusRepository() *SysRolesMenusRepository {
	SysRolesMenusRepo = &SysRolesMenusRepository{}
	return SysRolesMenusRepo
}

// getDB 获取数据库连接
func (r *SysRolesMenusRepository) getDB() *gorm.DB {
	return GetDB().Model(&model.SysRolesMenus{})
}

// BatchCreate 批量创建角色菜单关联
func (r *SysRolesMenusRepository) BatchCreate(ctx context.Context, associations []*model.SysRolesMenus) error {
	if len(associations) == 0 {
		return fmt.Errorf("associations list cannot be empty")
	}

	return r.getDB().WithContext(ctx).Create(associations).Error
}

// BatchDeleteByRoleID 批量删除指定角色的菜单关联
func (r *SysRolesMenusRepository) BatchDeleteByRoleID(ctx context.Context, roleID int64) error {
	return r.getDB().WithContext(ctx).Where("role_id = ?", roleID).Delete(&model.SysRolesMenus{}).Error
}

// BatchDeleteByMenuID 批量删除指定菜单的角色关联
func (r *SysRolesMenusRepository) BatchDeleteByMenuID(ctx context.Context, menuIDs []int64) error {
	return r.getDB().WithContext(ctx).Where("menu_id IN ?", menuIDs).Delete(&model.SysRolesMenus{}).Error
}

// BatchDeleteByRoleMenuIDs 批量删除指定角色菜单对的关联
func (r *SysRolesMenusRepository) BatchDeleteByRoleMenuIDs(ctx context.Context, associations []*model.SysRolesMenus) error {
	if len(associations) == 0 {
		return fmt.Errorf("associations list cannot be empty")
	}

	// 提取所有角色菜单对
	var conditions []map[string]interface{}
	for _, assoc := range associations {
		conditions = append(conditions, map[string]interface{}{
			"role_id": assoc.RoleID,
			"menu_id": assoc.MenuID,
		})
	}

	// 批量删除
	return r.getDB().WithContext(ctx).Where(conditions).Delete(&model.SysRolesMenus{}).Error
}

// BatchFindByRoleID 批量查询指定角色的菜单关联
func (r *SysRolesMenusRepository) BatchFindByRoleID(ctx context.Context, roleIDs []int64) ([]*model.SysRolesMenus, error) {
	if len(roleIDs) == 0 {
		return []*model.SysRolesMenus{}, nil
	}

	var associations []*model.SysRolesMenus
	err := r.getDB().WithContext(ctx).Where("role_id IN ?", roleIDs).Find(&associations).Error
	if err != nil {
		return nil, err
	}
	return associations, nil
}

// BatchFindByMenuID 批量查询指定菜单的角色关联
func (r *SysRolesMenusRepository) BatchFindByMenuID(ctx context.Context, menuID int64) ([]*model.SysRolesMenus, error) {
	var associations []*model.SysRolesMenus
	err := r.getDB().WithContext(ctx).Where("menu_id = ?", menuID).Find(&associations).Error
	if err != nil {
		return nil, err
	}
	return associations, nil
}

// InitTable 初始化表结构
func (r *SysRolesMenusRepository) InitTable() error {
	mod := &model.SysRolesMenus{}
	if err := r.getDB().AutoMigrate(mod); err != nil {
		return fmt.Errorf("failed to migrate table %s: %v", mod.TableName(), err)
	}
	return nil
}
