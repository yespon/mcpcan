package mysql

import (
	"context"
	"fmt"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"gorm.io/gorm"
)

var SysUsersRolesRepo *SysUsersRolesRepository

// SysUsersRolesRepository 用户角色关联仓库
type SysUsersRolesRepository struct{}

// NewSysUsersRolesRepository 创建用户角色关联仓库实例
func NewSysUsersRolesRepository() *SysUsersRolesRepository {
	SysUsersRolesRepo = &SysUsersRolesRepository{}
	return SysUsersRolesRepo
}

// getDB 获取数据库连接
func (r *SysUsersRolesRepository) getDB() *gorm.DB {
	return GetDB().Model(&model.SysUsersRoles{})
}

// BatchCreate 批量创建用户角色关联
func (r *SysUsersRolesRepository) BatchCreate(ctx context.Context, associations []*model.SysUsersRoles) error {
	if len(associations) == 0 {
		return fmt.Errorf("associations list cannot be empty")
	}

	return r.getDB().WithContext(ctx).Create(associations).Error
}

// BatchDeleteByUserID 批量删除指定用户的角色关联
func (r *SysUsersRolesRepository) BatchDeleteByUserID(ctx context.Context, userIDs []uint) error {
	return r.getDB().WithContext(ctx).Where("user_id IN ?", userIDs).Delete(&model.SysUsersRoles{}).Error
}

// BatchDeleteByRoleID 批量删除指定角色的用户关联
func (r *SysUsersRolesRepository) BatchDeleteByRoleID(ctx context.Context, roleID int64) error {
	return r.getDB().WithContext(ctx).Where("role_id = ?", roleID).Delete(&model.SysUsersRoles{}).Error
}

// BatchDeleteByUserRoleIDs 批量删除指定用户角色对的关联
func (r *SysUsersRolesRepository) BatchDeleteByUserRoleIDs(ctx context.Context, associations []*model.SysUsersRoles) error {
	if len(associations) == 0 {
		return fmt.Errorf("associations list cannot be empty")
	}

	// 提取所有用户角色对
	var conditions []map[string]interface{}
	for _, assoc := range associations {
		conditions = append(conditions, map[string]interface{}{
			"user_id": assoc.UserID,
			"role_id": assoc.RoleID,
		})
	}

	// 批量删除
	return r.getDB().WithContext(ctx).Where(conditions).Delete(&model.SysUsersRoles{}).Error
}

// BatchFindByUserID 批量查询指定用户的角色关联
func (r *SysUsersRolesRepository) BatchFindByUserID(ctx context.Context, userIDs []uint) ([]*model.SysUsersRoles, error) {
	var associations []*model.SysUsersRoles
	err := r.getDB().WithContext(ctx).Where("user_id IN ?", userIDs).Find(&associations).Error
	if err != nil {
		return nil, err
	}
	return associations, nil
}

// BatchFindByRoleID 批量查询指定角色的用户关联
func (r *SysUsersRolesRepository) BatchFindByRoleID(ctx context.Context, roleIDs []uint) ([]*model.SysUsersRoles, error) {
	var associations []*model.SysUsersRoles
	err := r.getDB().WithContext(ctx).Where("role_id IN ?", roleIDs).Find(&associations).Error
	if err != nil {
		return nil, err
	}
	return associations, nil
}

// InitTable 初始化表结构
func (r *SysUsersRolesRepository) InitTable() error {
	mod := &model.SysUsersRoles{}
	if err := r.getDB().AutoMigrate(mod); err != nil {
		return fmt.Errorf("failed to migrate table %s: %v", mod.TableName(), err)
	}
	return nil
}
