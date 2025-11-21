package task

import (
	"context"

	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// Task defines the interface for a background task.
type Task interface {
	Run(ctx context.Context) error
}

// StartTasks starts all the registered background tasks.
func StartTasks(logger *zap.Logger) {
	c := cron.New()

	// Register container monitor task to run every 5 minutes.
	containerMonitorTask := NewContainerMonitor(mysql.NewMcpInstanceRepository(), logger)
	c.AddFunc("*/5 * * * *", func() {
		if err := containerMonitorTask.Run(context.Background()); err != nil {
			logger.Error("Container monitor task failed", zap.Error(err))
		}
	})

	// Register gateway log cleaning task to run every 30 minutes.
	cleanGatewayLogTask := NewCleanGatewayLog(mysql.NewGatewayLogRepository(), logger)
	c.AddFunc("*/30 * * * *", func() {
		if err := cleanGatewayLogTask.Run(context.Background()); err != nil {
			logger.Error("Clean gateway log task failed", zap.Error(err))
		}
	})

	c.Start()
	logger.Info("Cron tasks started")
}
