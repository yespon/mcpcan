package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"gorm.io/gorm"
)

var SysUserRepo *SysUserRepository

// SysUserRepository 封装 sys_user 表的增删改查操作
type SysUserRepository struct{}

// NewSysUserRepository 创建 SysUserRepository 实例
func NewSysUserRepository() *SysUserRepository {
	SysUserRepo = &SysUserRepository{}
	return SysUserRepo
}

// getDB 获取数据库连接
func (r *SysUserRepository) getDB() *gorm.DB {
	return GetDB().Model(&model.SysUser{})
}

// Create 创建用户
func (r *SysUserRepository) Create(ctx context.Context, user *model.SysUser) error {
	now := time.Now()
	user.CreateTime = &now
	user.UpdateTime = &now
	return r.getDB().WithContext(ctx).Create(user).Error
}

// Update 更新用户
func (r *SysUserRepository) Update(ctx context.Context, user *model.SysUser) error {
	now := time.Now()
	user.UpdateTime = &now
	return r.getDB().WithContext(ctx).Where("user_id = ?", user.UserID).Save(user).Error
}

// Delete 删除用户
func (r *SysUserRepository) Delete(ctx context.Context, userID uint) error {
	return r.getDB().WithContext(ctx).Where("user_id = ?", userID).Delete(&model.SysUser{}).Error
}

// BatchDelete 批量删除用户
func (r *SysUserRepository) BatchDelete(ctx context.Context, userIDs []uint) error {
	return r.getDB().WithContext(ctx).Where("user_id IN ?", userIDs).Delete(&model.SysUser{}).Error
}

// FindByIds 根据用户ID列表查找用户
func (r *SysUserRepository) FindByIds(ctx context.Context, userIDs []uint) ([]*model.SysUser, error) {
	var users []*model.SysUser
	err := r.getDB().WithContext(ctx).Where("user_id IN ?", userIDs).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// FindByID 根据ID查找用户
func (r *SysUserRepository) FindByID(ctx context.Context, userID uint) (*model.SysUser, error) {
	var user model.SysUser
	err := r.getDB().WithContext(ctx).Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *SysUserRepository) FindByUsername(ctx context.Context, username string) (*model.SysUser, error) {
	var user model.SysUser
	err := r.getDB().WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByDeptID 根据部门ID查找用户
func (r *SysUserRepository) FindByDeptID(ctx context.Context, deptID []uint) ([]*model.SysUser, error) {
	var users []*model.SysUser
	err := r.getDB().WithContext(ctx).
		Where("dept_id IN ?", deptID).
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// FindWithPagination 分页查询用户（支持名称或邮箱的模糊查询，支持状态查询）
func (r *SysUserRepository) FindWithPagination(ctx context.Context, page, pageSize int, keyword string, enabled *bool, depIDs []uint, ids []uint) ([]*model.SysUser, int64, error) {
	var users []*model.SysUser
	var total int64

	query := r.getDB().WithContext(ctx)

	// 名称或邮箱的模糊查询
	if keyword != "" {
		query = query.Where("nick_name LIKE ? OR email LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 状态查询
	if enabled != nil {
		query = query.Where("enabled = ?", enabled)
	}

	if depIDs != nil {
		query = query.Where("dept_id IN (?)", depIDs)
	}
	// 用户ID查询
	if ids != nil {
		query = query.Where("user_id IN (?)", ids)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("user_id ASC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// InitTable 初始化表结构
func (r *SysUserRepository) InitTable() error {
	mod := &model.SysUser{}
	if err := r.getDB().AutoMigrate(mod); err != nil {
		return fmt.Errorf("failed to migrate table %s: %v", mod.TableName(), err)
	}
	return nil
}
