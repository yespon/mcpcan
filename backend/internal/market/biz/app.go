package biz

import (
	"context"
	"fmt"
	"sync"

	"github.com/kymo-mcp/mcpcan/internal/market/config"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
)

type App struct {
	config          *config.Config
	CodePackageList []*model.McpCodePackage
	McpTemplateList []*model.McpTemplate
	AdminUser       *model.SysUser
}

func NewApp(cfg *config.Config) *App {
	return &App{
		config: cfg,
	}
}

func (a *App) Initialize(ctx context.Context) error {
	// 1. Copy initial data (static assets)
	if err := a.copyInitData("./init-data/static", a.config.Storage.StaticPath); err != nil {
		return fmt.Errorf("failed to copy static data: %w", err)
	}

	// 2. Load admin user and menus if not in Kymo mode
	if a.config.RunMode != common.RunModeKymo {
		adminUser, err := a.createAdminUser()
		if err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}
		if err := a.createAdminRoleMenus(adminUser); err != nil {
			return fmt.Errorf("failed to create admin menus: %w", err)
		}
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.initDataScope(ctx); err != nil {
			fmt.Printf("Failed to init data scope: %v\n", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		RunMigrations()
	}()

	wg.Wait()

	return nil
}

func (a *App) initDataScope(ctx context.Context) error {
	if err := a.initCodePackage(ctx); err != nil {
		return fmt.Errorf("failed to init code package: %w", err)
	}
	if err := a.initOpenapi(ctx); err != nil {
		return fmt.Errorf("failed to init openapi: %w", err)
	}
	if err := a.initIntelligentAccess(ctx); err != nil {
		return fmt.Errorf("failed to init intelligent access: %w", err)
	}
	if err := a.initMcpTemplateData(ctx); err != nil {
		return fmt.Errorf("failed to init mcp templates: %w", err)
	}
	return nil
}
