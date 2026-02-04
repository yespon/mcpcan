package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	pb "github.com/kymo-mcp/mcpcan/api/authz/menu"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	i18nresp "github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/kymo-mcp/mcpcan/pkg/menu"
)

// MenuService menu HTTP service
type MenuService struct{}

// NewMenuService creates menu service instance
func NewMenuService() *MenuService {
	return &MenuService{}
}

// GetDeptTree gets department tree
func (s *MenuService) GetMenuTree(c *gin.Context) {
	var req pb.GetMenuTreeRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	dbMenus, _, err := mysql.SysMenuRepo.FindWithPagination(c.Request.Context(), 1, 9999, "")
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("Failed to get menu tree: %v", err))
		return
	}

	var permissionKey = map[string]*model.SysMenu{}
	for _, menu := range dbMenus {
		if menu.GetPermission() != "" {
			permissionKey[menu.GetPermission()] = menu
		}
	}

	menuTree := menu.GetMenus()
	filteredTree := filterMenuTree(menuTree, permissionKey)

	c.JSON(http.StatusOK, filteredTree)
}

// 根据 permissionKey 过滤菜单树, 返回的是 pb.SysMenuTreeNode 类型，id 通过 map 的 key 去匹配设置
func filterMenuTree(menuTree []*menu.Menu, permissionKey map[string]*model.SysMenu) []*pb.SysMenuTreeNode {
	var result []*pb.SysMenuTreeNode

	for _, m := range menuTree {
		// 递归过滤子菜单
		filteredChildren := filterMenuTree(m.Children, permissionKey)

		// 检查当前菜单是否在权限映射中
		if sysMenu, ok := permissionKey[m.Permission]; ok {
			// 创建新的SysMenuTreeNode
			treeNode := &pb.SysMenuTreeNode{
				Id:         sysMenu.MenuID,
				Title:      m.Title,
				Path:       m.Path,
				EngTitle:   m.EngTitle,
				Type:       int64(m.Type),
				Sort:       int32(m.Sort),
				Children:   filteredChildren,
				Permission: m.Permission,
			}
			result = append(result, treeNode)
		} else if len(filteredChildren) > 0 {
			// 如果当前菜单不在权限映射中，但有子菜单，则只保留子菜单
			result = append(result, filteredChildren...)
		}
	}

	return result
}
