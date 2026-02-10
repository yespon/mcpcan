package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"gorm.io/gorm"
)

var SysMenuRepo *SysMenuRepository

// SysMenuRepository 封装 sys_menu 表的增删改查操作
type SysMenuRepository struct{}

// NewSysMenuRepository 创建 SysMenuRepository 实例
func NewSysMenuRepository() *SysMenuRepository {
	SysMenuRepo = &SysMenuRepository{}
	return SysMenuRepo
}

// getDB 获取数据库连接
func (r *SysMenuRepository) getDB() *gorm.DB {
	return GetDB().Model(&model.SysMenu{})
}

// Create 创建菜单
func (r *SysMenuRepository) Create(ctx context.Context, menu *model.SysMenu) error {
	now := time.Now()
	menu.CreateTime = &now
	menu.UpdateTime = &now
	return r.getDB().WithContext(ctx).Create(menu).Error
}

// Update 更新菜单
func (r *SysMenuRepository) Update(ctx context.Context, menu *model.SysMenu) error {
	now := time.Now()
	menu.UpdateTime = &now
	return r.getDB().WithContext(ctx).Where("menu_id = ?", menu.MenuID).Save(menu).Error
}

// Delete 删除菜单
func (r *SysMenuRepository) Delete(ctx context.Context, menuID int64) error {
	return r.getDB().WithContext(ctx).Where("menu_id = ?", menuID).Delete(&model.SysMenu{}).Error
}

// BatchDelete 批量删除菜单
func (r *SysMenuRepository) BatchDelete(ctx context.Context, menuIDs []int64) error {
	return r.getDB().WithContext(ctx).Where("menu_id IN ?", menuIDs).Delete(&model.SysMenu{}).Error
}

// BatchFindByID 批量查询菜单
func (r *SysMenuRepository) FindByIDs(ctx context.Context, menuIDs []int64) ([]*model.SysMenu, error) {
	if len(menuIDs) == 0 {
		return []*model.SysMenu{}, nil
	}
	var menus []*model.SysMenu
	err := r.getDB().WithContext(ctx).Where("menu_id IN ?", menuIDs).Find(&menus).Error
	if err != nil {
		return nil, err
	}
	return menus, nil
}

// FindByID 根据ID查找菜单
func (r *SysMenuRepository) FindByID(ctx context.Context, menuID int64) (*model.SysMenu, error) {
	var menu model.SysMenu
	err := r.getDB().WithContext(ctx).Where("menu_id = ?", menuID).First(&menu).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

// FindWithPagination 分页查询菜单（支持菜单名称模糊查询）
func (r *SysMenuRepository) FindWithPagination(ctx context.Context, page, pageSize int, keyword string) ([]*model.SysMenu, int64, error) {
	var menus []*model.SysMenu
	var total int64

	query := r.getDB().WithContext(ctx)

	if keyword != "" {
		query = query.Where("title LIKE ? OR eng_title LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("menu_sort ASC").Find(&menus).Error; err != nil {
		return nil, 0, err
	}

	return menus, total, nil
}

// InitTable 初始化表结构
func (r *SysMenuRepository) InitTable() error {
	mod := &model.SysMenu{}
	if err := r.getDB().AutoMigrate(mod); err != nil {
		return fmt.Errorf("failed to migrate table %s: %v", mod.TableName(), err)
	}
	return nil
}
