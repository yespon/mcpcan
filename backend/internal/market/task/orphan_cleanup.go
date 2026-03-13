package task

import (
	"context"
	"fmt"

	"github.com/kymo-mcp/mcpcan/internal/market/biz"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"

	"go.uber.org/zap"
)

// OrphanCleanupTask scans every active runtime environment, lists all containers
// managed by mcpcan (via "managed-by" label), and deletes any that no longer have
// a corresponding active DB instance record. This handles containers that escaped
// system control due to delete failures, restarts, or other edge cases.
type OrphanCleanupTask struct {
	instanceRepo *mysql.McpInstanceRepository
	logger       *zap.Logger
}

// NewOrphanCleanup creates a new orphan cleanup task.
func NewOrphanCleanup(instanceRepo *mysql.McpInstanceRepository, logger *zap.Logger) Task {
	return &OrphanCleanupTask{
		instanceRepo: instanceRepo,
		logger:       logger,
	}
}

// Run is the main entry point called by the scheduler.
func (oc *OrphanCleanupTask) Run(ctx context.Context) error {
	oc.logger.Info("Starting orphan container cleanup task")

	// 1. Build a set of all container names that are "owned" by an active DB instance.
	//    We include ALL container_status values so we never accidentally delete a
	//    container that is still tracked, regardless of its current lifecycle state.
	instances, err := oc.instanceRepo.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("orphan cleanup: failed to list DB instances: %w", err)
	}

	knownContainers := make(map[string]struct{}, len(instances))
	for _, inst := range instances {
		if inst.ContainerName != "" && inst.Status == model.InstanceStatusActive {
			knownContainers[inst.ContainerName] = struct{}{}
		}
	}

	// 2. Iterate over all environments and scan each runtime for managed containers.
	environments, err := biz.GEnvironmentBiz.ListAllEnvironments(ctx)
	if err != nil {
		return fmt.Errorf("orphan cleanup: failed to list environments: %w", err)
	}

	totalCleaned := 0
	for _, env := range environments {
		cleaned, err := oc.cleanEnvironment(ctx, env, knownContainers)
		if err != nil {
			// Log and continue — don't abort other environments on a single failure
			oc.logger.Error("Failed to clean orphans in environment",
				zap.Uint("environment_id", uint(env.ID)),
				zap.String("environment_name", env.Name),
				zap.Error(err))
		}
		totalCleaned += cleaned
	}

	oc.logger.Info("Orphan container cleanup task completed",
		zap.Int("total_cleaned", totalCleaned))
	return nil
}

// cleanEnvironment scans one environment and removes orphan containers.
// Returns the count of containers deleted.
func (oc *OrphanCleanupTask) cleanEnvironment(ctx context.Context, env *model.McpEnvironment, knownContainers map[string]struct{}) (int, error) {
	entry, err := biz.GContainerBiz.GetRuntimeEntry(ctx, uint(env.ID))
	if err != nil {
		return 0, fmt.Errorf("failed to get runtime entry for env %d: %w", env.ID, err)
	}

	containerManager := entry.GetContainerManager()
	serviceManager := entry.GetServiceManager()

	// List all containers that carry the "managed-by=mcpcan" label
	managedLabel := map[string]string{
		"managed-by": common.SourceServerName,
	}
	containers, err := containerManager.ListByLabel(ctx, managedLabel)
	if err != nil {
		return 0, fmt.Errorf("failed to list managed containers: %w", err)
	}

	cleaned := 0
	for _, c := range containers {
		if _, ok := knownContainers[c.Name]; ok {
			// Container is tracked in DB — leave it alone
			continue
		}

		// Orphan detected: delete container and its associated service
		oc.logger.Warn("Detected orphan container, deleting",
			zap.Uint("environment_id", uint(env.ID)),
			zap.String("container_name", c.Name))

		if err := containerManager.Delete(ctx, c.Name); err != nil {
			oc.logger.Error("Failed to delete orphan container",
				zap.String("container_name", c.Name),
				zap.Error(err))
			// Continue to try deleting the service even if container delete fails
		}

		// Derive the service name from the container name and attempt deletion.
		// This is best-effort — the service might already be gone.
		svcName := deriveServiceName(c.Name)
		if svcName != "" {
			if err := serviceManager.Delete(ctx, svcName); err != nil {
				oc.logger.Warn("Failed to delete orphan service (may not exist)",
					zap.String("service_name", svcName),
					zap.Error(err))
			}
		}

		cleaned++
	}

	return cleaned, nil
}

// deriveServiceName converts a container deployment name to its associated service name.
// Container name pattern: mcp-instance-<id>-container
// Service name pattern:   mcp-instance-<id>-service
func deriveServiceName(containerName string) string {
	const containerSuffix = "-container"
	const serviceSuffix = "-service"
	if len(containerName) > len(containerSuffix) &&
		containerName[len(containerName)-len(containerSuffix):] == containerSuffix {
		return containerName[:len(containerName)-len(containerSuffix)] + serviceSuffix
	}
	return ""
}
