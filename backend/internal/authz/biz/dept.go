package biz

import (
	"context"
	"fmt"

	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
)

// DeptData department data access layer
type DeptData struct {
	ctx  context.Context
	repo *mysql.SysDeptRepository
}

// NewDeptData creates department data access layer instance
func NewDeptData(ctx context.Context) *DeptData {
	return &DeptData{
		ctx:  ctx,
		repo: mysql.SysDeptRepo,
	}
}

// BatchDeleteDepts deletes departments by IDs
func (d *DeptData) BatchDeleteDepts(c context.Context, deptIDs []uint) error {
	subDepts, err := mysql.SysDeptRepo.FindByParentID(c, deptIDs)
	if err != nil {
		return fmt.Errorf("failed to find sub-departments: %v", err)
	}

	deleteIDs := make([]uint, 0, len(subDepts))
	for _, dept := range subDepts {
		deleteIDs = append(deleteIDs, dept.DeptID)
	}
	deleteIDs = append(deleteIDs, deptIDs...)

	users, err := mysql.SysUserRepo.FindByDeptID(c, deleteIDs)
	if err != nil {
		return fmt.Errorf("failed to find users by department ID: %v", err)
	}

	if len(users) > 0 {
		var errorInfo = ""
		for _, user := range users {
			errorInfo += fmt.Sprintf("user %v (ID: %d) department ID: %d, ", user.Username, user.UserID, user.DeptID)
		}
		return fmt.Errorf("department has associated users, cannot be deleted: %s", errorInfo)
	}

	return d.repo.BatchDelete(c, deleteIDs)
}
