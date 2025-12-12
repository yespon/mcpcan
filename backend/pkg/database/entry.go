package database

import (
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
)

// Init 初始化数据库连接
func Init(config *common.MySQLConfig) error {
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
		return err
	}
	return nil
}
