package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	i18nresp "github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/kymo-mcp/mcpcan/pkg/menu"
	"github.com/kymo-mcp/mcpcan/pkg/utils"
)

func RBACAuthPathMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从context中获取userId
		userInfo, err := utils.GetCurrentUser(c)
		if err != nil {
			common.GinError(c, i18nresp.CodeInternalError, err.Error())
			c.Abort()
			return
		}

		// 获取用户角色的菜单权限
		if len(userInfo.RoleIds) == 0 {
			common.GinError(c, i18nresp.CodeInternalError, "no permission")
			c.Abort()
			return
		}
		roleMenus, err := mysql.SysRolesMenusRepo.BatchFindByRoleID(c.Request.Context(), userInfo.RoleIds)
		if err != nil {
			common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("Failed to find role menus: %v", err))
			c.Abort()
			return
		}
		menuIds := []int64{}
		for _, m := range roleMenus {
			menuIds = append(menuIds, m.MenuID)
		}
		if len(menuIds) == 0 {
			common.GinError(c, i18nresp.CodeInternalError, "no permission")
			c.Abort()
			return
		}
		menus, err := mysql.SysMenuRepo.FindByIDs(c.Request.Context(), menuIds)
		if err != nil {
			common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("Failed to find role menus: %v", err))
			c.Abort()
			return
		}
		var permissions = map[string]*model.SysMenu{}
		for _, m := range menus {
			permissions[m.GetPermission()] = m
		}

		method := c.Request.Method
		if method == "OPTIONS" {
			c.Next()
			return
		}

		path := c.FullPath()
		// 根据当前请求 path + method 获取需要的 permission
		permission := menu.GetPathPermission(path, strings.ToUpper(method))
		// 没获取到 permission，直接放行
		if permission == nil {
			c.Next()
		} else {
			// 获取到了判定是否有其中一个权限，有则放行
			for _, p := range permission {
				if _, ok := permissions[p]; ok {
					c.Next()
					return
				}
			}

			// 都没有权限，返回错误
			common.GinError(c, i18nresp.CodeInternalError, "no permission")
			c.Abort()
			return
		}
	}
}
