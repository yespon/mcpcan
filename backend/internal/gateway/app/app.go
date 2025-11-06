package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kymo-mcp/mcpcan/internal/gateway/config"
	"github.com/kymo-mcp/mcpcan/pkg/database"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
	"github.com/kymo-mcp/mcpcan/pkg/redis"

	"go.uber.org/zap"
)

// App application structure
type App struct {
	// config configuration
	config *config.Config

	// logger logger
	logger *zap.Logger

	// httpServer HTTP server
	httpServer *http.Server

	// shutdownCtx shutdown context
	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc
}

// New creates new application instance
func New() (*App, error) {
	// Load global configuration
	err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize logger
	if err := logger.Init(config.GlobalConfig.Log.Level, config.GlobalConfig.Log.Format); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Print configuration info
	logger.Debug("Version info", zap.String("version", fmt.Sprintf("%+v", config.GlobalConfig.VersionInfo)))

	ctx, cancel := context.WithCancel(context.Background())
	return &App{
		config:         config.GlobalConfig,
		logger:         logger.L().Logger,
		shutdownCtx:    ctx,
		shutdownCancel: cancel,
	}, nil
}

// Initialize initializes all application components
func (a *App) Initialize() error {
	// Initialize database
	if err := database.Init(&a.config.Database); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize redis
	if err := redis.Init(&a.config.Database.Redis); err != nil {
		return fmt.Errorf("failed to initialize redis: %w", err)
	}

	// Use global database repository instance (already initialized in init)
	if mysql.McpInstanceRepo == nil {
		return fmt.Errorf("McpInstanceRepo not properly initialized, please check the database initialization flow")
	}

	// Initialize HTTP server
	if err := a.initializeHTTPServer(); err != nil {
		return fmt.Errorf("failed to initialize HTTP server: %w", err)
	}

	a.logger.Info("Application initialization completed")
	return nil
}

// initializeHTTPServer initializes HTTP server
func (a *App) initializeHTTPServer() error {
	// Initialize Gin engine
	r := NewServer()

	// Create HTTP server
	serverAddr := fmt.Sprintf(":%d", config.GlobalConfig.Server.HttpPort)
	a.httpServer = &http.Server{
		Addr:    serverAddr,
		Handler: r,
	}

	return nil
}

// Run runs the application
func (a *App) Run() error {
	// Start HTTP server
	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("HTTP server failed to start", zap.Error(err))
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

// Shutdown gracefully shuts down the application
func (a *App) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
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

	a.logger.Info("Application gracefully shut down")
	return nil
}
