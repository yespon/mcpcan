package mysql

import (
	"context"
	"fmt"
	"time"

	pb "github.com/kymo-mcp/mcpcan/api/authz/dept"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"gorm.io/gorm"
)

var SysDeptRepo *SysDeptRepository

// SysDeptRepository 封装 sys_dept 表的增删改查操作
type SysDeptRepository struct{}

// NewSysDeptRepository 创建 SysDeptRepository 实例
func NewSysDeptRepository() *SysDeptRepository {
	SysDeptRepo = &SysDeptRepository{}
	return SysDeptRepo
}

// getDB 获取数据库连接
func (r *SysDeptRepository) getDB() *gorm.DB {
	return GetDB().Model(&model.SysDept{})
}

// Create 创建部门
func (r *SysDeptRepository) Create(ctx context.Context, dept *model.SysDept) error {
	now := time.Now()
	dept.CreateTime = &now
	dept.UpdateTime = &now
	return r.getDB().WithContext(ctx).Create(dept).Error
}

// Update 更新部门
func (r *SysDeptRepository) Update(ctx context.Context, dept *model.SysDept) error {
	now := time.Now()
	dept.UpdateTime = &now
	return r.getDB().WithContext(ctx).Where("dept_id = ?", dept.DeptID).Save(dept).Error
}

// Delete 删除部门
func (r *SysDeptRepository) Delete(ctx context.Context, deptID uint) error {
	return r.getDB().WithContext(ctx).Where("dept_id = ?", deptID).Delete(&model.SysDept{}).Error
}

// BatchDelete 批量删除部门
func (r *SysDeptRepository) BatchDelete(ctx context.Context, deptIDs []uint) error {
	return r.getDB().WithContext(ctx).Where("dept_id IN ?", deptIDs).Delete(&model.SysDept{}).Error
}

// FindByIDs 根据ID查找部门
func (r *SysDeptRepository) FindByIDs(ctx context.Context, deptIDs []uint) ([]*model.SysDept, error) {
	var depts []*model.SysDept
	err := r.getDB().WithContext(ctx).Where("dept_id IN ?", deptIDs).Find(&depts).Error
	if err != nil {
		return nil, err
	}
	return depts, nil
}

// FindByID 根据ID查找部门
func (r *SysDeptRepository) FindByID(ctx context.Context, deptID uint) (*model.SysDept, error) {
	var dept model.SysDept
	err := r.getDB().WithContext(ctx).Where("dept_id = ?", deptID).First(&dept).Error
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

// FindByName 根据部门名称查找部门
func (r *SysDeptRepository) FindByName(ctx context.Context, name string) (*model.SysDept, error) {
	var dept model.SysDept
	err := r.getDB().WithContext(ctx).Where("name = ?", name).First(&dept).Error
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

// FindByParentID 根据父部门ID查找子部门
func (r *SysDeptRepository) FindByParentID(ctx context.Context, parentIDs []uint) ([]*model.SysDept, error) {
	var depts []*model.SysDept
	err := r.getDB().WithContext(ctx).Where("pid IN ?", parentIDs).Find(&depts).Error
	if err != nil {
		return nil, err
	}
	return depts, nil
}

// FindWithPagination 分页查询部门（支持部门名称模糊查询）
func (r *SysDeptRepository) FindWithPagination(ctx context.Context, page, pageSize int, keyword string, pid *uint, enabled int32) ([]*model.SysDept, int64, error) {
	var depts []*model.SysDept
	var total int64

	query := r.getDB().WithContext(ctx)

	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	// pid 不为0时，查询子部门
	if pid != nil {
		if *pid > 0 {
			query = query.Where("pid = ?", pid)
		}
		// pid 为0时，查询所有部门
	} else {
		// pid 为空的时候查询顶级
		query = query.Where("pid IS NULL")
	}

	if enabled > 0 {
		enable := true
		if pb.DeptStatus(enabled) == pb.DeptStatus_DeptStatusDisabled {
			enable = false
		}
		query = query.Where("enabled = ?", enable)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("dept_sort ASC").Find(&depts).Error; err != nil {
		return nil, 0, err
	}

	return depts, total, nil
}

// InitTable 初始化表结构
func (r *SysDeptRepository) InitTable() error {
	mod := &model.SysDept{}
	if err := r.getDB().AutoMigrate(mod); err != nil {
		return fmt.Errorf("failed to migrate table %s: %v", mod.TableName(), err)
	}
	return nil
}
