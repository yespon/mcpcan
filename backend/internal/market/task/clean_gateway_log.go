package task

import (
	"context"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"go.uber.org/zap"
)

// CleanGatewayLogImpl implements the gateway log cleaning task.
type CleanGatewayLogImpl struct {
	logRepo *mysql.GatewayLogRepository
	logger  *zap.Logger
}

// NewCleanGatewayLog creates a new CleanGatewayLog task.
func NewCleanGatewayLog(
	logRepo *mysql.GatewayLogRepository,
	logger *zap.Logger,
) Task {
	return &CleanGatewayLogImpl{
		logRepo: logRepo,
		logger:  logger,
	}
}

// Run executes the log cleaning task.
func (c *CleanGatewayLogImpl) Run(ctx context.Context) error {
	c.logger.Info("Starting gateway log cleaning task")

	cutoff := time.Now().Add(-24 * time.Hour)
	rowsAffected, err := c.logRepo.DeleteBefore(ctx, cutoff)
	if err != nil {
		c.logger.Error("Failed to clean gateway logs", zap.Error(err))
		return err
	}

	c.logger.Info("Gateway log cleaning task completed",
		zap.Int64("deleted_rows", rowsAffected),
		zap.Time("cutoff_time", cutoff),
	)

	return nil
}
