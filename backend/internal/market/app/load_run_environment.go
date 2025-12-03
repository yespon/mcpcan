package app

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
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
	applied, err := mysql.SysMigrationRepo.HasApplied(ctx, migrationName)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}
	if applied {
		return nil
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
	if err := mysql.SysMigrationRepo.Apply(ctx, migrationName); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return nil
}

func (a *App) createKubernetesEnv(ctx context.Context, cfg interface{}, creatorID string) (*model.McpEnvironment, error) {
	// Re-access config to be type-safe
	k8sCfg := a.config.RunEnvironment.Kubernetes
	name := a.config.RunEnvironment.Name
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

	namespace := "default"
	if ctxCfg, ok := clientConfig.Contexts[clientConfig.CurrentContext]; ok {
		if ctxCfg.Namespace != "" {
			namespace = ctxCfg.Namespace
		}
	}

	return &model.McpEnvironment{
		Name:        name,
		Environment: model.McpEnvironmentKubernetes,
		Config:      string(content),
		CreatorID:   creatorID,
		Namespace:   namespace,
	}, nil
}

func (a *App) createDockerEnv(ctx context.Context, cfg interface{}, creatorID string) (*model.McpEnvironment, error) {
	dockerCfg := a.config.RunEnvironment.Docker
	name := a.config.RunEnvironment.Name
	if name == "" {
		name = "Default-Run-Environment"
	}

	if dockerCfg.Host == "" {
		return nil, fmt.Errorf("docker host is empty")
	}

	// Validate TLS config if enabled
	var certPathForDB string
	if dockerCfg.UseTLS {
		if dockerCfg.CertPath == "" || dockerCfg.KeyPath == "" || dockerCfg.CAPath == "" {
			return nil, fmt.Errorf("docker UseTLS is true but certPath, keyPath, or caPath is missing")
		}

		// Verify files exist
		if _, err := os.Stat(dockerCfg.CertPath); err != nil {
			return nil, fmt.Errorf("certPath error: %v", err)
		}
		if _, err := os.Stat(dockerCfg.KeyPath); err != nil {
			return nil, fmt.Errorf("keyPath error: %v", err)
		}
		if _, err := os.Stat(dockerCfg.CAPath); err != nil {
			return nil, fmt.Errorf("caPath error: %v", err)
		}

		// Verify certificate validity
		if _, err := tls.LoadX509KeyPair(dockerCfg.CertPath, dockerCfg.KeyPath); err != nil {
			return nil, fmt.Errorf("failed to load x509 key pair: %v", err)
		}
		caPEM, err := os.ReadFile(dockerCfg.CAPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA cert: %v", err)
		}
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caPEM) {
			return nil, fmt.Errorf("failed to append CA cert")
		}

		// Check directory consistency for current Docker runtime implementation
		// DockerRuntime expects all certs in one directory and standard names (cert.pem, key.pem, ca.pem)
		// Check if they are in the same directory
		certDir := filepath.Dir(dockerCfg.CertPath)
		keyDir := filepath.Dir(dockerCfg.KeyPath)
		caDir := filepath.Dir(dockerCfg.CAPath)

		if certDir != keyDir || certDir != caDir {
			return nil, fmt.Errorf("docker certificates (cert, key, ca) must be in the same directory for the current runtime support")
		}

		// Check filenames (DockerRuntime expects cert.pem, key.pem, ca.pem)
		if filepath.Base(dockerCfg.CertPath) != "cert.pem" ||
			filepath.Base(dockerCfg.KeyPath) != "key.pem" ||
			filepath.Base(dockerCfg.CAPath) != "ca.pem" {
			return nil, fmt.Errorf("docker certificate files must be named cert.pem, key.pem, and ca.pem")
		}

		certPathForDB = certDir
	}

	// Prepare DB config
	dbConfig := model.DockerEnvironmentConfig{
		Host:     dockerCfg.Host,
		UseTLS:   dockerCfg.UseTLS,
		CertPath: certPathForDB,
		Network:  dockerCfg.Network,
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
	}, nil
}
