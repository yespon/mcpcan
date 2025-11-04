package config

import (
    "os"
    "strconv"
    "strings"
    "fmt"

    "github.com/kymo-mcp/mcpcan/pkg/common"
    "github.com/kymo-mcp/mcpcan/pkg/utils"
    "github.com/kymo-mcp/mcpcan/pkg/version"

	"github.com/spf13/viper"
)

var GlobalConfig *Config

// Config represents configuration structure
type Config struct {
    ServiceName string                `mapstructure:"-"`
    VersionInfo *version.VersionInfo  `mapstructure:"-"`
    Server      common.ServerConfig   `mapstructure:"server"`
    Services    common.Services       `mapstructure:"services"`
    Domain      string                `mapstructure:"domain"`
    Database    common.DatabaseConfig `mapstructure:"database"`
    Code        common.CodeConfig     `mapstructure:"code"`
    Market      common.MarketConfig   `mapstructure:"market"`
    Log         common.LogConfig      `mapstructure:"log"`
    Secret      string                `mapstructure:"secret"`
    Storage     common.StorageConfig  `mapstructure:"storage"`
    // RunMode indicates current running mode, e.g., "demo" or "prod"
    RunMode string `mapstructure:"runMode"`
    // DemoMaxInstances specifies the maximum number of active instances allowed in demo mode
    DemoMaxInstances int `mapstructure:"demoMaxInstances"`
}

var serviceName = "market"
var cfgFileName = "market.yaml"

// GetConfig gets global configuration
func GetConfig() *Config {
	return GlobalConfig
}

// Load loads configuration file
func Load() (*Config, error) {
    v := viper.New()
    v.SetConfigType("yaml")

	// If configuration file path is not specified, try to find it automatically
	var err error
	configPath, err := common.FindConfigFile(cfgFileName)
	if err != nil {
		return nil, err
	}

	// Set configuration file path
	v.SetConfigFile(configPath)

	// Read configuration file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// Parse configuration
	var config Config
    if err := v.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("failed to parse config file: %v", err)
    }

    // Read run mode from environment if not provided in config
    if strings.TrimSpace(config.RunMode) == "" {
        if envMode := strings.TrimSpace(os.Getenv("RUN_MODE")); envMode != "" {
            config.RunMode = envMode
        }
    }

    // Read demo max instances from environment if provided
    if envDemoMax := strings.TrimSpace(os.Getenv("DEMO_MAX_INSTANCES")); envDemoMax != "" {
        if v, err := strconv.Atoi(envDemoMax); err == nil && v > 0 {
            config.DemoMaxInstances = v
        }
    }

    if config.Code.Upload.MaxFileSize == 0 {
        config.Code.Upload.MaxFileSize = 100
    }

	if config.Code.Upload.AllowedExtensions == nil {
		config.Code.Upload.AllowedExtensions = []string{".zip", ".tar.gz", ".tar", ".rar"}
	}

	if config.Storage.RootPath == "" {
		config.Storage.RootPath = "/app/data"
	}
	utils.MkdirP(config.Storage.RootPath)

	if config.Storage.CodePath == "" {
		config.Storage.CodePath = "/app/data/code-package"
	}
	utils.MkdirP(config.Storage.CodePath)

	if config.Storage.StaticPath == "" {
		config.Storage.StaticPath = "/app/data/static"
	}
    utils.MkdirP(config.Storage.StaticPath)

    // Set sensible defaults
    if config.DemoMaxInstances <= 0 {
        config.DemoMaxInstances = 10
    }

    // Append Version information
    config.ServiceName = serviceName
    config.VersionInfo = version.GetVersionInfo()

    GlobalConfig = &config

    return &config, nil
}

// IsDemoMode returns true if the application is running in demo mode
func IsDemoMode() bool {
    if GlobalConfig == nil {
        return false
    }
    return strings.EqualFold(strings.TrimSpace(GlobalConfig.RunMode), "demo")
}

// GetDemoMaxInstances returns the maximum allowed active instances in demo mode
func GetDemoMaxInstances() int {
    if GlobalConfig == nil || GlobalConfig.DemoMaxInstances <= 0 {
        return 10
    }
    return GlobalConfig.DemoMaxInstances
}
