package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kymo-mcp/mcpcan/pkg/utils"
	"go.uber.org/zap"

	pb "github.com/kymo-mcp/mcpcan/api/authz/role"
	"github.com/kymo-mcp/mcpcan/internal/authz/biz"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	i18nresp "github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
)

// RoleService role HTTP service
type RoleService struct {
	roleBiz *biz.RoleData
}

// NewRoleService creates role service instance
func NewRoleService() *RoleService {
	return &RoleService{
		roleBiz: biz.NewRoleData(context.Background()),
	}
}

// CreateRole creates role
func (s *RoleService) CreateRole(c *gin.Context) {
	var req pb.CreateRoleRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}
	if req.Name == "" {
		common.GinError(c, i18nresp.CodeInternalError, "name is required")
		return
	}
	if req.Level <= 0 {
		common.GinError(c, i18nresp.CodeInternalError, "level is required")
		return
	}

	userInfo, err := utils.GetCurrentUser(c)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	var level = int(req.Level)

	// Convert request to model
	roleModel := &model.SysRole{
		Name:        req.Name,
		Description: &req.Description,
		Level:       &level,
		DataScope:   &req.DataScope,
		CreateBy:    &userInfo.Username,
		UpdateBy:    &userInfo.Username,
	}

	// Create role
	if err := mysql.SysRoleRepo.Create(c.Request.Context(), roleModel); err != nil {
		logger.Error("Failed to create role", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("Failed to create role: %v", err))
		return
	}

	// Convert model to response
	respData := s.convertModelToProto(roleModel)

	common.GinSuccess(c, respData)
}

// UpdateRole updates role
func (s *RoleService) UpdateRole(c *gin.Context) {
	var req pb.UpdateRoleRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}
	if req.Name == "" {
		common.GinError(c, i18nresp.CodeInternalError, "name is required")
		return
	}
	if req.Level <= 0 {
		common.GinError(c, i18nresp.CodeInternalError, "level is required")
		return
	}

	userInfo, err := utils.GetCurrentUser(c)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	// Get role by ID
	roleModel, err := mysql.SysRoleRepo.FindByID(c.Request.Context(), uint(req.Id))
	if err != nil {
		logger.Error("Failed to get role", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("Failed to get role: %v", err))
		return
	}

	var level = int(req.Level)
	roleModel.Name = req.Name
	roleModel.Description = &req.Description
	roleModel.Level = &level
	roleModel.DataScope = &req.DataScope
	roleModel.UpdateBy = &userInfo.Username

	// Update role
	if err := mysql.SysRoleRepo.Update(c.Request.Context(), roleModel); err != nil {
		logger.Error("Failed to update role", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("Failed to update role: %v", err))
		return
	}

	// Convert model to response
	respData := s.convertModelToProto(roleModel)

	common.GinSuccess(c, respData)
}

// DeleteRole deletes role
func (s *RoleService) DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, "Invalid role ID")
		return
	}

	// Delete role
	if err := s.roleBiz.BatchDeleteRoles(c.Request.Context(), []uint{uint(id)}); err != nil {
		logger.Error("Failed to delete role", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("Failed to delete role: %v", err))
		return
	}

	common.GinSuccess(c, nil)
}

// BatchDeleteRole batch deletes roles
func (s *RoleService) BatchDeleteRole(c *gin.Context) {

	var req pb.BatchDeleteRolesRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}
	// Convert IDs to uint slice
	var roleIDs []uint
	for _, id := range req.Ids {
		roleIDs = append(roleIDs, uint(id))
	}

	// Batch delete roles
	if err := mysql.SysRoleRepo.BatchDelete(c.Request.Context(), roleIDs); err != nil {
		logger.Error("Failed to batch delete roles", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("Failed to batch delete roles: %v", err))
		return
	}

	common.GinSuccess(c, nil)
}

// ListRoles gets role list with pagination
func (s *RoleService) ListRoles(c *gin.Context) {
	var req pb.ListRolesRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Set default page and page size if not provided
	page := int(req.Page)
	if page <= 0 {
		page = 1
	}

	pageSize := int(req.Size)
	if pageSize <= 0 {
		pageSize = 10
	}

	// Use blurry query as keyword
	keyword := req.Blurry

	// Get roles with pagination
	roles, total, err := mysql.SysRoleRepo.FindWithPagination(c.Request.Context(), page, pageSize, keyword, nil)
	if err != nil {
		logger.Error("Failed to get roles with pagination", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("Failed to get roles: %v", err))
		return
	}

	// Convert models to response
	var respRoles []*pb.SysRole
	for _, role := range roles {
		respRoles = append(respRoles, s.convertModelToProto(role))
	}

	common.GinSuccess(c, &pb.ListRolesResponse{
		Total: total,
		List:  respRoles,
	})
}

// convertModelToProto converts model.SysRole to pb.SysRole
func (s *RoleService) convertModelToProto(role *model.SysRole) *pb.SysRole {

	resp := &pb.SysRole{
		Id:        int64(role.RoleID),
		Name:      role.Name,
		CreatedAt: 0,
		UpdatedAt: 0,
	}

	if role.Level != nil {
		level := int32(*role.Level)
		resp.Level = int64(level)
	}

	// Set description if exists
	if role.Description != nil {
		resp.Description = *role.Description
	}

	// Set created time if exists
	if role.CreateTime != nil {
		resp.CreatedAt = role.CreateTime.Unix()
	}

	// Set updated time if exists
	if role.UpdateTime != nil {
		resp.UpdatedAt = role.UpdateTime.Unix()
	}

	// Set created by if exists
	if role.CreateBy != nil {
		resp.CreatedBy = *role.CreateBy
	}

	// Set updated by if exists
	if role.UpdateBy != nil {
		resp.UpdatedBy = *role.UpdateBy
	}

	return resp
}

// SaveRoleMenus saves role menus
func (s *RoleService) SaveRoleMenus(c *gin.Context) {
	var req pb.SaveRoleMenusRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	err := mysql.SysRolesMenusRepo.BatchDeleteByRoleID(c.Request.Context(), req.RoleId)
	if err != nil {
		logger.Error("Failed to batch delete role menus", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("Failed to batch delete role menus: %v", err))
		return
	}

	associations := []*model.SysRolesMenus{}
	for _, menuID := range req.MenuIds {
		associations = append(associations, &model.SysRolesMenus{
			RoleID: req.RoleId,
			MenuID: menuID,
		})
	}

	err = mysql.SysRolesMenusRepo.BatchCreate(c.Request.Context(), associations)
	if err != nil {
		logger.Error("Failed to batch create role menus", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("Failed to batch create role menus: %v", err))
		return
	}

	common.GinSuccess(c, nil)
}

// FindRoleMenus finds role menus
func (s *RoleService) FindRoleMenus(c *gin.Context) {
	var req pb.FindRoleMenusRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	roleMenus, err := mysql.SysRolesMenusRepo.BatchFindByRoleID(c.Request.Context(), req.RoleIds)
	if err != nil {
		logger.Error("Failed to find role menus", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("Failed to find role menus: %v", err))
		return
	}

	// Convert models to response
	menuIds := []int64{}
	var menuRoleId = make(map[int64]int64)
	for _, menu := range roleMenus {
		menuIds = append(menuIds, menu.MenuID)
		menuRoleId[menu.MenuID] = menu.RoleID
	}

	menus, err := mysql.SysMenuRepo.FindByIDs(c.Request.Context(), menuIds)
	if err != nil {
		logger.Error("Failed to find role menus", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("Failed to find role menus: %v", err))
		return
	}

	sysRoleMenus := []*pb.SysRoleMenu{}
	var permissions []string
	for _, menu := range menus {
		permissions = append(permissions, menu.GetPermission())
		sysRoleMenus = append(sysRoleMenus, &pb.SysRoleMenu{
			Id:         menu.MenuID,
			Name:       menu.GetTitle(),
			Permission: menu.GetPermission(),
			RoleId:     menuRoleId[menu.MenuID],
			Type:       menu.GetType(),
			EngName:    menu.GetEngTitle(),
			Sort:       menu.GetMenuSort(),
			Path:       menu.GetPath(),
		})
	}

	common.GinSuccess(c, &pb.FindRoleMenusResponse{
		Menus:       sysRoleMenus,
		Permissions: permissions,
	})
}
