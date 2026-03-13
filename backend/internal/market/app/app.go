package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/kymo-mcp/mcpcan/internal/market/biz"
	cfg "github.com/kymo-mcp/mcpcan/internal/market/config"
	"github.com/kymo-mcp/mcpcan/internal/market/service"
	"github.com/kymo-mcp/mcpcan/internal/market/task"

	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
	"github.com/kymo-mcp/mcpcan/pkg/middleware"
	"github.com/kymo-mcp/mcpcan/pkg/redis"
	"github.com/kymo-mcp/mcpcan/pkg/scheduler"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// App application structure
type App struct {
	// config configuration
	config *cfg.Config

	// logger logger
	logger *zap.Logger

	// scheduler scheduler
	scheduler scheduler.Scheduler

	// taskManager task manager
	taskManager task.TaskManager

	// httpServer HTTP server
	httpServer *http.Server

	// ginEngine Gin engine
	ginEngine *gin.Engine

	// bizApp business logic app
	bizApp *biz.App

	// shutdownCtx shutdown context
	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc
}

// New creates new application instance
func New() (*App, error) {
	// Load global configuration
	cfg, err := cfg.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	// Initialize logger
	if err := logger.Init(cfg.Log.Level, cfg.Log.Format); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Print configuration information
	logger.Debug("Version info", zap.String("version", fmt.Sprintf("%+v", cfg.VersionInfo)))

	ctx, cancel := context.WithCancel(context.Background())
	return &App{
		config:         cfg,
		logger:         logger.L().Logger,
		shutdownCtx:    ctx,
		shutdownCancel: cancel,
	}, nil
}

// Initialize initialize all application components
func (a *App) Initialize() error {

	// Initialize database
	if err := a.loadMysql(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize business application (includes migrated init service logic)
	a.bizApp = biz.NewApp(a.config)
	if err := a.bizApp.Initialize(context.Background()); err != nil {
		return fmt.Errorf("failed to initialize business application: %w", err)
	}

	// Register enterprise plugins (no-op when enterprise features are disabled)
	if err := a.registerEnterprisePlugin(); err != nil {
		return err
	}

	// Initialize Redis
	if err := redis.Init(&a.config.Database.Redis); err != nil {
		return fmt.Errorf("failed to initialize Redis: %w", err)
	}

	// load run environment
	if err := a.initRunEnvironment(context.Background()); err != nil {
		return fmt.Errorf("failed to init run environment: %w", err)
	}

	// Use global database repository instance (already initialized in init)
	if mysql.McpInstanceRepo == nil {
		return fmt.Errorf("McpInstanceRepo not properly initialized, please check database initialization process")
	}

	// Start scheduler
	if err := a.initializeScheduler(); err != nil {
		return fmt.Errorf("failed to initialize scheduler: %w", err)
	}

	// Initialize task manager, no longer depends on global container runtime
	a.taskManager = task.NewTaskManager(
		mysql.McpInstanceRepo,
		a.scheduler,
		a.logger,
	)

	// Set up global tasks
	if err := a.taskManager.SetupGlobalTasks(a.shutdownCtx); err != nil {
		return fmt.Errorf("failed to set up global tasks: %w", err)
	}

	// Initialize HTTP server
	if err := a.initializeHTTPServer(); err != nil {
		return fmt.Errorf("failed to initialize HTTP server: %w", err)
	}

	// 用配置中的 staticPath 重新初始化 AiFileManager（覆盖 init() 中的默认路径）
	biz.GAiFileManager = biz.NewAiFileManager(a.config.Storage.StaticPath)

	a.logger.Info("Application initialization completed")
	return nil
}

// loadMysql initializes MySQL database connection and loads necessary tables
func (a *App) loadMysql() error {
	tableInitializers := []func() (string, error){
		func() (string, error) {
			repo := mysql.NewMcpCodePackageRepository()
			return (&model.McpCodePackage{}).TableName(), repo.InitTable()
		},
		func() (string, error) {
			repo := mysql.NewMcpEnvironmentRepository()
			return (&model.McpEnvironment{}).TableName(), repo.InitTable()
		},
		func() (string, error) {
			repo := mysql.NewGatewayLogRepository()
			return (&model.GatewayLog{}).TableName(), repo.InitTable()
		},
		func() (string, error) {
			repo := mysql.NewMcpInstanceRepository()
			return (&model.McpInstance{}).TableName(), repo.InitTable()
		},
		func() (string, error) {
			repo := mysql.NewMcpMigrationRepository()
			return (&model.Migration{}).TableName(), repo.InitTable()
		},
		func() (string, error) {
			repo := mysql.NewMcpOpenapiPackageRepository()
			return (&model.McpOpenapiPackage{}).TableName(), repo.InitTable()
		},
		func() (string, error) {
			repo := mysql.NewMcpTemplateRepository()
			return (&model.McpTemplate{}).TableName(), repo.InitTable()
		},
		func() (string, error) {
			repo := mysql.NewMcpToIntelligentTaskRepository()
			return (&model.McpToIntelligentTask{}).TableName(), repo.InitTable()
		},
		func() (string, error) {
			repo := mysql.NewMcpToIntelligentTaskLogRepository()
			return (&model.McpToIntelligentTaskLog{}).TableName(), repo.InitTable()
		},
		func() (string, error) {
			repo := mysql.NewMcpTokenRepository()
			return (&model.McpToken{}).TableName(), repo.InitTable()
		},
		func() (string, error) {
			repo := mysql.NewAiSessionRepository()
			return (&model.AiSession{}).TableName(), repo.InitTable()
		},
		func() (string, error) {
			repo := mysql.NewAiMessageRepository()
			return (&model.AiMessage{}).TableName(), repo.InitTable()
		},
		func() (string, error) {
			repo := mysql.NewAiModelAccessRepository()
			return (&model.AiModelAccess{}).TableName(), repo.InitTable()
		},
		func() (string, error) {
			if a.config.RunMode == common.RunModeKymo {
				model.SetIntelligentAccessTableName("intelligent_access")
				mod := &model.IntelligentAccess{}
				return mod.TableName(), nil
			} else {
				model.SetIntelligentAccessTableName("mcpcan_intelligent_access")
				mod := &model.IntelligentAccess{}
				repo := mysql.NewIntelligentAccessRepository()
				return mod.TableName(), repo.InitTable()
			}
		},
	}

	// Sys tables - only load when NOT in kymo mode
	if a.config.RunMode != common.RunModeKymo {
		tableInitializers = append(tableInitializers,
			func() (string, error) {
				repo := mysql.NewSysDeptRepository()
				return (&model.SysDept{}).TableName(), repo.InitTable()
			},
			func() (string, error) {
				repo := mysql.NewSysRoleRepository()
				return (&model.SysRole{}).TableName(), repo.InitTable()
			},
			func() (string, error) {
				repo := mysql.NewSysUserRepository()
				return (&model.SysUser{}).TableName(), repo.InitTable()
			},
			func() (string, error) {
				repo := mysql.NewSysUsersRolesRepository()
				return (&model.SysUsersRoles{}).TableName(), repo.InitTable()
			},
			func() (string, error) {
				repo := mysql.NewSysMenuRepository()
				return (&model.SysMenu{}).TableName(), repo.InitTable()
			},
			func() (string, error) {
				repo := mysql.NewSysRolesMenusRepository()
				return (&model.SysRolesMenus{}).TableName(), repo.InitTable()
			},
		)
	}

	// Append enterprise table initializers (empty when enterprise features are disabled)
	tableInitializers = append(tableInitializers, a.enterpriseTableInitializers()...)

	return database.Init(&a.config.Database.MySQL, tableInitializers...)
}

// initializeScheduler initialize scheduler
func (a *App) initializeScheduler() error {
	globalScheduler := scheduler.GetGlobalScheduler()
	if globalScheduler == nil {
		return fmt.Errorf("global scheduler not initialized")
	}

	a.scheduler = globalScheduler.GetTaskManager().GetScheduler()
	a.logger.Info("Scheduler initialized successfully")

	return nil
}

// initializeHTTPServer initialize HTTP server
func (a *App) initializeHTTPServer() error {

	a.ginEngine = gin.Default()

	// Set up middleware
	a.setupMiddleware()

	// Initialize Gin engine
	a.setupHttpServer()

	// Create HTTP server
	serverAddr := fmt.Sprintf(":%d", a.config.Server.HttpPort)
	a.httpServer = &http.Server{
		Addr:    serverAddr,
		Handler: a.ginEngine,
	}

	return nil
}

// Run run application
func (a *App) Run() error {
	// Start task manager
	err := a.taskManager.StartMonitoring(a.shutdownCtx)
	if err != nil {
		return fmt.Errorf("failed to start task manager: %w", err)
	}

	// Start HTTP server
	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("HTTP server startup failed", zap.Error(err))
		}
	}()

	a.logger.Info("Application started successfully",
		zap.String("address", a.httpServer.Addr))

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.logger.Info("Shutting down application...")

	// Graceful shutdown
	return a.Shutdown()
}

// Shutdown gracefully shutdown application
func (a *App) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Stop task manager
	if a.taskManager != nil {
		err := a.taskManager.StopMonitoring(ctx)
		if err != nil {
			a.logger.Error("Failed to stop task manager", zap.Error(err))
		}
	}

	// Close HTTP server
	if a.httpServer != nil {
		if err := a.httpServer.Shutdown(ctx); err != nil {
			a.logger.Error("HTTP server shutdown failed", zap.Error(err))
			return err
		}
	}

	// Cancel application context
	if a.shutdownCancel != nil {
		a.shutdownCancel()
	}

	a.logger.Info("Application has been gracefully shut down")
	return nil
}

// setupHttpServer initialize Gin engine and register all routes
func (a *App) setupHttpServer() {
	// Set file upload size limit, default is 32 MiB, according to configuration file set to 100 MiB
	a.ginEngine.MaxMultipartMemory = int64(a.config.Code.Upload.MaxFileSize) << 20

	// Get route prefix
	routerPrefix := common.GetMarketRoutePrefix()
	routerPrefix = strings.Trim(routerPrefix, "/")

	// Register instance management interface
	instanceService := service.NewInstanceService()
	a.ginEngine.POST(fmt.Sprintf("/%s/instance/create", routerPrefix), instanceService.CreateHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/instance/:instanceId", routerPrefix), instanceService.DetailHandler)
	a.ginEngine.PUT(fmt.Sprintf("/%s/instance/edit", routerPrefix), instanceService.EditHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/instance/list", routerPrefix), instanceService.ListHandler)
	a.ginEngine.PUT(fmt.Sprintf("/%s/instance/disabled", routerPrefix), instanceService.DisabledHandler)
	a.ginEngine.PUT(fmt.Sprintf("/%s/instance/restart", routerPrefix), instanceService.RestartHandler)
	a.ginEngine.DELETE(fmt.Sprintf("/%s/instance/:instanceId", routerPrefix), instanceService.DeleteHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/instance/status/:instanceId", routerPrefix), instanceService.StatusHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/instance/logs", routerPrefix), instanceService.LogsHandler)
	a.ginEngine.PUT(fmt.Sprintf("/%s/instance/token/control", routerPrefix), instanceService.TokenControlHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/instance/token/list", routerPrefix), instanceService.TokenListByInstanceIDHandler)
	a.ginEngine.PUT(fmt.Sprintf("/%s/instance/token/edit", routerPrefix), instanceService.TokenEditHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/instance/token/delete", routerPrefix), instanceService.TokenDeleteHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/instance/openapi/create", routerPrefix), instanceService.CreateOpenapiHandler)
	a.ginEngine.PUT(fmt.Sprintf("/%s/instance/openapi/edit", routerPrefix), instanceService.UpdateOpenapiHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/instance/list-tools", routerPrefix), instanceService.ListToolsHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/instance/call-tools", routerPrefix), instanceService.CallToolHandler)

	// Create resource management service instance
	resourceService := service.NewResourceService()
	a.ginEngine.GET(fmt.Sprintf("/%s/resources/pvcs", routerPrefix), resourceService.ListPVCsHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/resources/pvcs", routerPrefix), resourceService.CreatePVCHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/resources/nodes", routerPrefix), resourceService.ListNodesHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/resources/storage-classes", routerPrefix), resourceService.ListStorageClassesHandler)
	// Register Docker volume management interface
	a.ginEngine.GET(fmt.Sprintf("/%s/resources/docker/volumes", routerPrefix), resourceService.ListDockerVolumesHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/resources/docker/volumes/create", routerPrefix), resourceService.CreateDockerVolumeHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/resources/docker/volumes/find", routerPrefix), resourceService.FindDockerVolumeHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/resources/docker/volumes/remove", routerPrefix), resourceService.RemoveDockerVolumeHandler)

	// Create environment management service instance
	environmentService := service.NewEnvironmentService()
	a.ginEngine.GET(fmt.Sprintf("/%s/environments", routerPrefix), environmentService.ListEnvironmentsHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/environments/namespaces", routerPrefix), environmentService.ListNamespacesHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/environments/:id/test", routerPrefix), environmentService.TestConnectivityHandler)

	// Register code management interface
	codeService := service.NewCodeService()
	a.ginEngine.POST(fmt.Sprintf("/%s/code/upload", routerPrefix), codeService.UploadPackage)
	a.ginEngine.GET(fmt.Sprintf("/%s/code/tree", routerPrefix), codeService.GetCodeTree)
	a.ginEngine.GET(fmt.Sprintf("/%s/code/get", routerPrefix), codeService.GetCodeFile)
	a.ginEngine.POST(fmt.Sprintf("/%s/code/edit", routerPrefix), codeService.EditCodeFile)
	a.ginEngine.GET(fmt.Sprintf("/%s/code/download/:packageId", routerPrefix), codeService.DownloadPackage)
	a.ginEngine.GET(fmt.Sprintf("/%s/code/packages", routerPrefix), codeService.GetCodePackageList)
	a.ginEngine.DELETE(fmt.Sprintf("/%s/code/packages/:packageId", routerPrefix), codeService.DeleteCodePackage)

	// Register OpenAPI document management interface
	openapiService := service.NewOpenapiService()
	a.ginEngine.POST(fmt.Sprintf("/%s/openapi/upload", routerPrefix), openapiService.UploadOpenapiFile)
	a.ginEngine.GET(fmt.Sprintf("/%s/openapi/content", routerPrefix), openapiService.GetOpenapiFileContent)
	a.ginEngine.POST(fmt.Sprintf("/%s/openapi/edit", routerPrefix), openapiService.EditOpenapiFile)
	a.ginEngine.GET(fmt.Sprintf("/%s/openapi/download/:openapiFileId", routerPrefix), openapiService.DownloadOpenapiFile)
	a.ginEngine.GET(fmt.Sprintf("/%s/openapi/files", routerPrefix), openapiService.GetOpenapiFileList)
	a.ginEngine.DELETE(fmt.Sprintf("/%s/openapi/files/:openapiFileId", routerPrefix), openapiService.DeleteOpenapiFile)

	// Register template management interface
	templateService := service.NewTemplateService(context.Background())
	a.ginEngine.POST(fmt.Sprintf("/%s/template/create", routerPrefix), templateService.TemplateCreateHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/template/:templateId", routerPrefix), templateService.TemplateDetailHandler)
	a.ginEngine.PUT(fmt.Sprintf("/%s/template/edit", routerPrefix), templateService.TemplateEditHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/template/list", routerPrefix), templateService.TemplateListHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/template/list/pagination", routerPrefix), templateService.TemplateListWithPaginationHandler)
	a.ginEngine.DELETE(fmt.Sprintf("/%s/template/:templateId", routerPrefix), templateService.TemplateDeleteHandler)

	// Register market management interface
	marketService := service.NewMarketService()
	if marketService != nil {
		a.ginEngine.POST(fmt.Sprintf("/%s/market/list", routerPrefix), marketService.ListMarketServices)
		a.ginEngine.GET(fmt.Sprintf("/%s/market/detail", routerPrefix), marketService.GetMarketServiceDetail)
		a.ginEngine.GET(fmt.Sprintf("/%s/market/category", routerPrefix), marketService.GetMarketCategories)
		a.ginEngine.GET(fmt.Sprintf("/%s/market/config", routerPrefix), marketService.GetMarketConfig)
	}

	// Register platform market management interface
	platformMarketService := service.NewPlatformMarketService()
	if platformMarketService != nil {
		a.ginEngine.GET(fmt.Sprintf("/%s/platform/list", routerPrefix), platformMarketService.ListMcpServer)
	}

	// Register Traefik Gateway interfaces
	gatewayService := service.NewGatewayService()
	a.ginEngine.GET(fmt.Sprintf("/%s/gateway/auth", routerPrefix), gatewayService.AuthHandler)
	a.ginEngine.Any(fmt.Sprintf("%s/*any", common.GatewayRoutePrefix), gatewayService.ProxyHandler)

	// Register gateway log interface
	gatewayLogService := service.NewGatewayLogService()
	a.ginEngine.POST(fmt.Sprintf("/%s/gateway-log/find", routerPrefix), gatewayLogService.FindHandler)

	// Register storage management interface
	storageService := service.NewStorageService(context.Background())
	a.ginEngine.POST(fmt.Sprintf("/%s/storage/image", routerPrefix), storageService.UploadImageHandler)

	// Register dashboard management interface
	dashboardService := service.NewDashboardService(context.Background())
	a.ginEngine.GET(fmt.Sprintf("/%s/dashboard/statistical", routerPrefix), dashboardService.StatisticalHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/dashboard/available-cases", routerPrefix), dashboardService.AvailableCasesHandler)

	// Register intelligent access management interface
	intelligentAccessService := service.NewIntelligentAccessService(context.Background())
	a.ginEngine.POST(fmt.Sprintf("/%s/intelligent_access", routerPrefix), intelligentAccessService.CreateHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/intelligent_access/list", routerPrefix), intelligentAccessService.ListHandler)
	a.ginEngine.DELETE(fmt.Sprintf("/%s/intelligent_access/delete", routerPrefix), intelligentAccessService.DeleteHandler)
	a.ginEngine.PUT(fmt.Sprintf("/%s/intelligent_access/edit", routerPrefix), intelligentAccessService.UpdateHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/intelligent_access/test-connection", routerPrefix), intelligentAccessService.TestConnectionHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/intelligent_access/list-user-space", routerPrefix), intelligentAccessService.ListUserSpaceHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/intelligent_access/install-n8n-plugin", routerPrefix), intelligentAccessService.InstallN8NPluginHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/intelligent_access/check-n8n", routerPrefix), intelligentAccessService.CheckN8NHandler)

	mcpToIntelligentTaskService := service.NewMcpToIntelligentTaskService(context.Background())
	// Register mcp to intelligent task management interface
	a.ginEngine.POST(fmt.Sprintf("/%s/mcp_to_intelligent_task", routerPrefix), mcpToIntelligentTaskService.CreateHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/mcp_to_intelligent_task/:id", routerPrefix), mcpToIntelligentTaskService.GetHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/mcp_to_intelligent_task/list", routerPrefix), mcpToIntelligentTaskService.ListHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/mcp_to_intelligent_task/:id/cancel", routerPrefix), mcpToIntelligentTaskService.CancelHandler)

	// Register AI session management interface
	aiSessionService := service.NewAiSessionService(context.Background())
	a.ginEngine.POST(fmt.Sprintf("/%s/ai/sessions", routerPrefix), aiSessionService.CreateHandler)
	a.ginEngine.PUT(fmt.Sprintf("/%s/ai/sessions", routerPrefix), aiSessionService.UpdateHandler)
	a.ginEngine.DELETE(fmt.Sprintf("/%s/ai/sessions/:id", routerPrefix), aiSessionService.DeleteHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/ai/sessions/:id", routerPrefix), aiSessionService.GetHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/ai/sessions", routerPrefix), aiSessionService.ListHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/ai/sessions/:id/messages", routerPrefix), aiSessionService.GetSessionMessagesHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/ai/sessions/:id/usage", routerPrefix), aiSessionService.GetSessionUsageHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/ai/sessions/:id/chat", routerPrefix), aiSessionService.ChatHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/ai/sessions/reset-memory", routerPrefix), aiSessionService.ResetMemoryHandler)

	// Register AI model access management interface
	aiModelAccessService := service.NewAiModelAccessService(context.Background())
	a.ginEngine.POST(fmt.Sprintf("/%s/ai/models", routerPrefix), aiModelAccessService.CreateHandler)
	a.ginEngine.PUT(fmt.Sprintf("/%s/ai/models", routerPrefix), aiModelAccessService.UpdateHandler)
	a.ginEngine.DELETE(fmt.Sprintf("/%s/ai/models/:id", routerPrefix), aiModelAccessService.DeleteHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/ai/models/:id", routerPrefix), aiModelAccessService.GetHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/ai/models", routerPrefix), aiModelAccessService.ListHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/ai/models/available", routerPrefix), aiModelAccessService.GetAvailableModelsHandler)
	a.ginEngine.GET(fmt.Sprintf("/%s/ai/models/supported", routerPrefix), aiModelAccessService.GetSupportedModelsHandler)
	a.ginEngine.POST(fmt.Sprintf("/%s/ai/models/test", routerPrefix), aiModelAccessService.TestConnectionHandler)

	// Register file upload interface
	a.ginEngine.POST(fmt.Sprintf("/%s/ai/files/upload", routerPrefix), aiSessionService.UploadFileHandler)

	// Register enterprise routes (no-op when enterprise features are disabled)
	a.registerEnterpriseRoutes(a.ginEngine, routerPrefix)

	// Health check
	a.ginEngine.GET("/health", func(c *gin.Context) {
		i18n.SuccessResponse(c, gin.H{"status": "ok"})
	})
}

// setupMiddleware set up middleware
func (a *App) setupMiddleware() {
	// Add panic recovery middleware
	a.ginEngine.Use(middleware.PanicRecovery())

	// Add CORS middleware (must be early to handle OPTIONS)
	a.ginEngine.Use(middleware.CORSMiddleware())

	// Add request response logging middleware
	a.ginEngine.Use(middleware.RequestResponseLoggingMiddleware())

	// Add internationalization middleware
	a.ginEngine.Use(middleware.I18nMiddleware())

	// Add security middleware
	a.ginEngine.Use(middleware.SecurityMiddleware(a.config.Secret))

	// // set user info to context
	a.ginEngine.Use(middleware.AppendUserMiddleware())

	// // Add RBAC middleware
	// a.ginEngine.Use(middleware.RBACAuthPathMiddleware())

	// Add error handling middleware (must be last)
	a.ginEngine.Use(middleware.ErrorHandler())

	// Set custom error handler
	a.ginEngine.NoRoute(middleware.NotFoundHandler)
	a.ginEngine.NoMethod(middleware.MethodNotAllowedHandler)
}
