package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/kymo-mcp/mcpcan/internal/authz/config"
	"github.com/kymo-mcp/mcpcan/internal/authz/service"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	dbpkg "github.com/kymo-mcp/mcpcan/pkg/database"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
	"github.com/kymo-mcp/mcpcan/pkg/middleware"
	"github.com/kymo-mcp/mcpcan/pkg/redis"
)

// App application structure
type App struct {
	config     *config.Config
	logger     *zap.Logger
	httpServer *http.Server
	ginEngine  *gin.Engine
}

// New creates application instance
func New() *App {
	// Load configuration
	if err := config.Load(); err != nil {
		return nil
	}

	// Initialize logger
	if err := logger.Init(config.GlobalConfig.Log.Level, config.GlobalConfig.Log.Format); err != nil {
		return nil
	}

	// Print configuration information
	logger.Debug("Version info", zap.String("version", fmt.Sprintf("%+v", config.GlobalConfig.VersionInfo)))

	return &App{
		config: config.GlobalConfig,
		logger: logger.L().Logger,
	}
}

// Initialize initializes the application
func (a *App) Initialize() error {
	// Initialize Redis
	if err := redis.Init(&config.GlobalConfig.Database.Redis); err != nil {
		return fmt.Errorf("failed to initialize redis: %w", err)
	}

	// Initialize database
	if err := a.loadMysql(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Register enterprise plugins
	if err := a.registerEnterprisePlugin(); err != nil {
		return fmt.Errorf("failed to register enterprise plugin: %w", err)
	}

	// Setup HTTP server
	if err := a.setupHTTPServer(); err != nil {
		return fmt.Errorf("failed to setup HTTP server: %w", err)
	}

	logger.Info("Authz service initialized successfully")
	return nil
}

// loadMysql initializes MySQL database connection and loads necessary tables
func (a *App) loadMysql() error {
	var tableInitializers []func() (string, error)
	// 当运行模式不是 kymo 时，加载逻辑中使用的表
	if a.config.RunMode != "kymo" {
		tableInitializers = []func() (string, error){
			func() (string, error) {
				mysql.NewSysUserRepository()
				return (&model.SysUser{}).TableName(), nil
			},
			func() (string, error) {
				mysql.NewSysRoleRepository()
				return (&model.SysRole{}).TableName(), nil
			},
			func() (string, error) {
				mysql.NewSysDeptRepository()
				return (&model.SysDept{}).TableName(), nil
			},
			func() (string, error) {
				mysql.NewSysUsersRolesRepository()
				return (&model.SysUsersRoles{}).TableName(), nil
			},
			func() (string, error) {
				mysql.NewSysMenuRepository()
				return (&model.SysMenu{}).TableName(), nil
			},
			func() (string, error) {
				mysql.NewSysRolesMenusRepository()
				return (&model.SysRolesMenus{}).TableName(), nil
			},
		}
	}

	return dbpkg.Init(&a.config.Database.MySQL, tableInitializers...)
}

// setupHTTPServer sets up HTTP server
func (a *App) setupHTTPServer() error {
	// Set Gin mode
	gin.SetMode(gin.DebugMode)

	// Create Gin engine
	a.ginEngine = gin.New()

	// Setup middleware
	a.setupMiddleware()

	// Setup routes
	a.setupRoutes()

	// Create HTTP server
	a.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", a.config.Server.HttpPort),
		Handler: a.ginEngine,
	}

	logger.Info("HTTP server setup completed")
	return nil
}

// setupMiddleware sets up middleware
func (a *App) setupMiddleware() {
	// Add middleware
	a.ginEngine.Use(gin.Recovery())
	a.ginEngine.Use(middleware.RequestResponseLoggingMiddleware())

	// Add security middleware
	a.ginEngine.Use(middleware.SecurityMiddleware(config.GlobalConfig.Secret))
	// set user info to context
	a.ginEngine.Use(middleware.AppendUserMiddleware())
	// Add RBAC middleware
	a.ginEngine.Use(middleware.RBACAuthPathMiddleware())
}

// setupRoutes sets up routes
func (a *App) setupRoutes() {
	// Create service instances
	userAuthService := service.NewUserAuthService()

	// Health check
	a.ginEngine.GET("/health", func(c *gin.Context) {
		common.GinSuccess(c, map[string]string{"status": "ok"})
	})

	// API version prefix
	authzGroup := a.ginEngine.Group(common.GetAuthzRoutePrefix())

	// If running in kymo mode, only load the validation interface
	if a.config.RunMode == common.RunModeKymo {
		authzGroup.GET("/validate", userAuthService.KymoValidateToken)
		return
	}

	deptService := service.NewDeptService()
	deptGroup := authzGroup.Group("/depts")
	{
		deptGroup.POST("", deptService.CreateDept)
		deptGroup.PUT("", deptService.UpdateDept)
		deptGroup.DELETE("/:id", deptService.DeleteDept)
		deptGroup.DELETE("/batch-delete", deptService.BatchDeleteDept)
		deptGroup.PUT("/:id/status", deptService.UpdateDeptStatus)
		deptGroup.GET("/list-depts", deptService.FindDepts)
		deptGroup.GET("/tree", deptService.GetDeptTree)
	}

	roleService := service.NewRoleService()
	roleGroup := authzGroup.Group("/roles")
	{
		roleGroup.POST("", roleService.CreateRole)
		roleGroup.PUT("", roleService.UpdateRole)
		roleGroup.DELETE("/:id", roleService.DeleteRole)
		roleGroup.DELETE("/batch-delete", roleService.BatchDeleteRole)
		roleGroup.GET("/list-roles", roleService.ListRoles)
		roleGroup.POST("/save-menus", roleService.SaveRoleMenus)
		roleGroup.POST("/find-menus", roleService.FindRoleMenus)
	}

	userService := service.NewUserService()
	userGroup := authzGroup.Group("/users")
	userGroup.Use(middleware.AuthTokenMiddleware(config.GlobalConfig.Secret))
	{
		userGroup.POST("", userService.CreateUser)
		userGroup.PUT("/:id", userService.UpdateUser)
		userGroup.DELETE("/:id", userService.DeleteUser)
		userGroup.DELETE("/batch-delete", userService.BatchDelete)
		userGroup.GET("/list-users", userService.ListUsers)
		userGroup.PUT("/update-password", userService.UpdatePassword)
		userGroup.PUT("/update-avatar", userService.UpdateAvatar)
		userGroup.POST("/add-role", userService.AddUserRole)
		userGroup.POST("/remove-role", userService.RemoveUserRole)
	}

	menuService := service.NewMenuService()
	menuGroup := authzGroup.Group("/menus")
	{
		menuGroup.GET("/tree", menuService.GetMenuTree)
	}

	{
		// User login
		authzGroup.POST("/login", userAuthService.Login)

		// User logout
		authzGroup.POST("/logout", userAuthService.Logout)

		// Get user configuration
		authzGroup.GET("/user-info", userAuthService.GetUserInfo)

		// Refresh Token
		authzGroup.POST("/refresh", userAuthService.RefreshToken)

		// Validate Token
		authzGroup.GET("/validate", userAuthService.ValidateToken)

		// Get encryption key
		authzGroup.POST("/encryption-key", userAuthService.GetEncryptionKey)
	}
}

// Run runs the application
func (a *App) Run() error {
	// Start HTTP server
	logger.Info("Starting HTTP server", zap.Int("port", a.config.Server.HttpPort))
	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down authz service...")

	// Graceful shutdown
	return a.Shutdown()
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if a.httpServer != nil {
		logger.Info("Shutting down HTTP server...")
		if err := a.httpServer.Shutdown(ctx); err != nil {
			logger.Error("Failed to shutdown HTTP server", zap.Error(err))
		} else {
			logger.Info("HTTP server stopped gracefully")
		}
	}

	// Close database connection
	if err := mysql.Close(); err != nil {
		logger.Error("Failed to close database", zap.Error(err))
		return err
	}

	logger.Info("Authz service shutdown completed")
	return nil
}
