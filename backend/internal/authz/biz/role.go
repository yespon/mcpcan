package biz

import (
	"context"
	"fmt"

	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
)

// RoleData role data access layer
type RoleData struct {
	ctx  context.Context
	repo *mysql.SysRoleRepository
}

// NewRoleData creates role data access layer instance
func NewRoleData(ctx context.Context) *RoleData {
	return &RoleData{
		ctx:  ctx,
		repo: mysql.SysRoleRepo,
	}
}

// BatchDeleteRoles deletes roles by IDs
func (d *RoleData) BatchDeleteRoles(ctx context.Context, roleIDs []uint) error {
	associations, err := mysql.SysUsersRolesRepo.BatchFindByRoleID(ctx, roleIDs)
	if err != nil {
		return err
	}

	userIds := make([]uint, 0, len(associations))
	for _, association := range associations {
		userIds = append(userIds, association.UserID)
	}

	users, err := mysql.SysUserRepo.FindByIds(ctx, userIds)
	if err != nil {
		return err
	}

	if len(users) > 0 {
		var errorInfo = ""
		for _, user := range users {
			errorInfo += fmt.Sprintf("user %v (ID: %d), ", user.Username, user.UserID)
		}
		return fmt.Errorf("role has associated users, cannot be deleted: %s", errorInfo)
	}

	return d.repo.BatchDelete(ctx, roleIDs)
}
