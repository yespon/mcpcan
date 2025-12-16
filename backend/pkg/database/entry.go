package database

import (
	"time"

	"go.uber.org/zap"

	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
)

// Init 初始化数据库连接
func Init(config *common.MySQLConfig, initializers ...func() (string, error)) error {
	// 初始化 MySQL 配置
	mysqlConfig := &mysql.Config{
		Host:                config.Host,
		Port:                config.Port,
		Username:            config.Username,
		Password:            config.Password,
		Database:            config.Database,
		ConnectTimeout:      20 * time.Second,
		MaxIdleConns:        50,
		MaxOpenConns:        300,
		HealthCheckInterval: 30 * time.Second,
		MaxRetries:          12,
		RetryInterval:       5 * time.Second,
	}
	err := mysql.InitDB(mysqlConfig)
	if err != nil {
		logger.Error("Failed to connect to database", zap.Error(err))
		return err
	}

	if len(initializers) > 0 {
		logger.Info("Starting database table initialization", zap.Int("count", len(initializers)))
		for i, init := range initializers {
			tableName, err := init()
			if err != nil {
				logger.Error("Failed to initialize table",
					zap.Int("index", i),
					zap.String("table", tableName),
					zap.Error(err))
				return err
			}
			if tableName != "" {
				logger.Info("Table initialized successfully", zap.String("table", tableName))
			}
		}
		logger.Info("Database table initialization completed successfully")
	}

	return nil
}
