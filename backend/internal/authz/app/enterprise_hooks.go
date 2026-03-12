package app

import (
	"fmt"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	enterpriseplugin "github.com/kymo-mcp/mcpcan/pkg/enterprise.ee/database/plugin"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
)

// registerEnterprisePlugin registers enterprise GORM plugins (e.g. data permission).
func (a *App) registerEnterprisePlugin() error {
	if a.config.CodeMode != common.EnterpriseCodeCodeMode {
		return nil
	}
	if err := mysql.GetDB().Use(&enterpriseplugin.GlobalDataPermissionPlugin{}); err != nil {
		return fmt.Errorf("failed to register data permission plugin: %w", err)
	}
	return nil
}
