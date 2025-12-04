package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"gorm.io/gorm"
	"k8s.io/client-go/tools/clientcmd"
)

// initRunEnvironment initializes the default run environment from config
func (a *App) initRunEnvironment(ctx context.Context) error {
	cfg := a.config.RunEnvironment

	if !cfg.Enabled {
		return nil
	}

	// Check if migration already applied
	migrationName := "init_run_environment"
	_, err := mysql.McpMigrationRepo.FindByName(ctx, migrationName)
	if err == nil {
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	fmt.Println("Initializing default run environment...")

	var env *model.McpEnvironment
	var validationErr error

	// Determine creator ID (default to 1 for system admin)
	creatorID := "1"
	// Try to find admin user (optional, just to be sure)
	adminUser, err := mysql.SysUserRepo.FindByID(ctx, 1)
	if err == nil && adminUser != nil {
		creatorID = strconv.FormatUint(uint64(adminUser.UserID), 10)
	}

	switch strings.ToLower(cfg.Type) {
	case "kubernetes":
		env, validationErr = a.createKubernetesEnv(ctx, cfg, creatorID)
	case "docker":
		env, validationErr = a.createDockerEnv(ctx, cfg, creatorID)
	default:
		return fmt.Errorf("unsupported environment type: %s", cfg.Type)
	}

	if validationErr != nil {
		// Fatal error as requested
		fmt.Printf("Run environment configuration error: %v\n", validationErr)
		os.Exit(1)
	}

	if env != nil {
		// Create in DB
		// Check if name exists first to avoid duplicate error if migration check failed but data exists
		existing, err := mysql.McpEnvironmentRepo.FindByName(ctx, env.Name)
		if err == nil && existing != nil {
			fmt.Printf("Environment '%s' already exists, skipping creation.\n", env.Name)
		} else {
			if err := mysql.McpEnvironmentRepo.Create(ctx, env); err != nil {
				return fmt.Errorf("failed to create environment in db: %w", err)
			}
			fmt.Printf("Created default run environment: %s (ID: %d)\n", env.Name, env.ID)
		}
	}

	// Record migration
	migration := &model.Migration{
		Name:        migrationName,
		CompletedAt: time.Now(),
	}
	if err := mysql.McpMigrationRepo.Create(ctx, migration); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return nil
}

func (a *App) createKubernetesEnv(ctx context.Context, cfg common.RunEnvironmentConfig, creatorID string) (*model.McpEnvironment, error) {
	// Re-access config to be type-safe
	k8sCfg := cfg.Kubernetes
	name := cfg.Name
	if name == "" {
		name = "Default-Run-Environment"
	}

	if k8sCfg.ConfigPath == "" {
		return nil, fmt.Errorf("kubernetes configPath is empty")
	}

	// Read kubeconfig
	content, err := os.ReadFile(k8sCfg.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read kubeconfig file '%s': %v", k8sCfg.ConfigPath, err)
	}

	// Validate kubeconfig format and extract namespace
	clientConfig, err := clientcmd.Load(content)
	if err != nil {
		return nil, fmt.Errorf("invalid kubeconfig format: %v", err)
	}

	// Determine namespace: Config > Kubeconfig > Default
	namespace := k8sCfg.Namespace
	if namespace == "" {
		if ctxCfg, ok := clientConfig.Contexts[clientConfig.CurrentContext]; ok {
			if ctxCfg.Namespace != "" {
				namespace = ctxCfg.Namespace
			}
		}
	}
	if namespace == "" {
		return nil, fmt.Errorf("kubernetes namespace is empty")
	}

	return &model.McpEnvironment{
		Name:        name,
		Environment: model.McpEnvironmentKubernetes,
		Config:      string(content),
		CreatorID:   creatorID,
		Namespace:   namespace,
		Level:       model.McpEnvironmentLevelSystem,
	}, nil
}

func (a *App) createDockerEnv(ctx context.Context, cfg common.RunEnvironmentConfig, creatorID string) (*model.McpEnvironment, error) {
	dockerCfg := cfg.Docker
	name := cfg.Name
	if name == "" {
		name = "Default-Run-Environment"
	}

	if dockerCfg.Host == "" {
		return nil, fmt.Errorf("docker host is empty")
	}
	if dockerCfg.Network == "" {
		return nil, fmt.Errorf("docker network is empty")
	}

	// If host is a unix socket, verify it exists
	if strings.HasPrefix(dockerCfg.Host, "unix://") {
		socketPath := strings.TrimPrefix(dockerCfg.Host, "unix://")
		if _, err := os.Stat(socketPath); err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("docker socket file does not exist: %s", socketPath)
			}
			return nil, fmt.Errorf("failed to check docker socket file: %v", err)
		}
	} else {
		return nil, fmt.Errorf("docker host must be a unix socket")
	}

	// Note: We do not support loading TLS certificates from file configuration anymore.
	// Config file only supports local docker.sock. Non-local environments must be configured via system backend.

	// Prepare DB config
	dbConfig := model.DockerEnvironmentConfig{
		Host:    dockerCfg.Host,
		Network: dockerCfg.Network,
		UseTLS:  false, // Always false for file-based init as per requirement
	}

	configBytes, err := json.Marshal(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal docker config: %v", err)
	}

	return &model.McpEnvironment{
		Name:        name,
		Environment: model.McpEnvironmentDocker,
		Config:      string(configBytes),
		CreatorID:   creatorID,
		Level:       model.McpEnvironmentLevelSystem,
	}, nil
}
