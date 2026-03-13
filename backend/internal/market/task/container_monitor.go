package task

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kymo-mcp/mcpcan/internal/market/biz"
	"github.com/kymo-mcp/mcpcan/pkg/container"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"

	"go.uber.org/zap"
)

// orphanRecreateThreshold is the number of consecutive GetInfo failures before
// we consider the container truly gone and trigger a recreate.
const orphanRecreateThreshold = 3

// ContainerMonitorImpl container monitoring implementation
type ContainerMonitorImpl struct {
	// instanceRepo instance database operations
	instanceRepo *mysql.McpInstanceRepository

	// logger logger
	logger *zap.Logger

	// maxConcurrency maximum concurrent check count
	maxConcurrency int

	// failureCount tracks consecutive GetInfo failures per instance (in-memory, resets on restart)
	failureCount   map[string]int
	failureCountMu sync.Mutex
}

// NewContainerMonitor creates a new container monitor
func NewContainerMonitor(
	instanceRepo *mysql.McpInstanceRepository,
	logger *zap.Logger,
) Task {
	return &ContainerMonitorImpl{
		instanceRepo:   instanceRepo,
		logger:         logger,
		maxConcurrency: 10,
		failureCount:   make(map[string]int),
	}
}

// Run monitors all containers
func (cm *ContainerMonitorImpl) Run(ctx context.Context) error {
	cm.logger.Info("Starting global container monitoring task")

	// Get hosting instances in service
	instances, err := cm.instanceRepo.FindHostingInstances(ctx)
	if err != nil {
		cm.logger.Error("Failed to get MCP instances with specified container status", zap.Error(err))
		return fmt.Errorf("failed to get MCP instances with specified container status: %w", err)
	}

	cm.logger.Info("Retrieved MCP instances with specified container status",
		zap.Int("count", len(instances)),
		zap.Strings("statuses", []string{string(model.ContainerStatusPending), string(model.ContainerStatusRunning)}))

	// Use concurrent checking of container status, check at most 10 at the same time
	semaphore := make(chan struct{}, cm.maxConcurrency)
	var wg sync.WaitGroup

	// Used to collect errors
	errorChan := make(chan error, len(instances))

	// Concurrently check instance container status
	for _, instance := range instances {
		wg.Add(1)
		go func(inst *model.McpInstance) {
			defer wg.Done()

			// Get semaphore to control concurrency
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			err := cm.CheckContainer(ctx, inst)
			if err != nil {
				cm.logger.Error("Container check failed",
					zap.String("instance_id", inst.InstanceID),
					zap.String("container_name", inst.ContainerName),
					zap.Error(err))
				// Send error to error channel, but don't block
				select {
				case errorChan <- err:
				default:
				}
			}
		}(instance)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errorChan)

	// Collect and log errors, but don't interrupt the entire monitoring process
	errorCount := 0
	for err := range errorChan {
		errorCount++
		if errorCount == 1 {
			cm.logger.Warn("Errors occurred during container checking", zap.Error(err))
		}
	}

	if errorCount > 0 {
		cm.logger.Warn("Container checking completed, some instance checks failed",
			zap.Int("total_instances", len(instances)),
			zap.Int("failed_count", errorCount))
	}

	cm.logger.Info("Global container monitoring task execution completed")
	return nil
}

// CheckContainer checks a single container
func (cm *ContainerMonitorImpl) CheckContainer(ctx context.Context, instance *model.McpInstance) error {
	cm.logger.Debug("Starting container check",
		zap.String("instance_id", instance.InstanceID))

	// If instance has no container name, it means no container has been created yet, skip
	if instance.ContainerName == "" {
		cm.logger.Debug("Instance has not created container yet, skipping check",
			zap.String("instance_id", instance.InstanceID))
		return nil
	}

	// Get creation parameters
	containerCreateOptions := &container.ContainerCreateOptions{}
	if err := json.Unmarshal([]byte(instance.ContainerCreateOptions), containerCreateOptions); err != nil {
		cm.logger.Error("Failed to deserialize creation parameters",
			zap.String("instance_id", instance.InstanceID),
			zap.Error(err))
		return err
	}

	// Get runtime entry
	entry, err := biz.GContainerBiz.GetRuntimeEntry(ctx, instance.EnvironmentID)
	if err != nil {
		return err
	}

	return cm.checkContainerStatus(ctx, instance, entry, containerCreateOptions)
}

// checkContainerStatus checks container status (supports both K8s and Docker)
func (cm *ContainerMonitorImpl) checkContainerStatus(ctx context.Context, instance *model.McpInstance,
	entry container.Runtime, containerCreateOptions *container.ContainerCreateOptions) error {

	// Get current timestamp (milliseconds)
	currentTime := time.Now().UnixMilli()

	containerManager := entry.GetContainerManager()
	// Get container detailed information
	containerInfo, err := containerManager.GetInfo(ctx, instance.ContainerName)
	if err != nil {
		// D: Backoff protection — only recreate after reaching consecutive failure threshold.
		// This prevents transient network errors / API timeouts from triggering spurious recreations.
		cm.failureCountMu.Lock()
		cm.failureCount[instance.InstanceID]++
		count := cm.failureCount[instance.InstanceID]
		cm.failureCountMu.Unlock()

		cm.logger.Warn("GetInfo failed for container",
			zap.String("instance_id", instance.InstanceID),
			zap.String("container_name", instance.ContainerName),
			zap.Int("consecutive_failures", count),
			zap.Int("threshold", orphanRecreateThreshold),
			zap.Error(err))

		if count < orphanRecreateThreshold {
			// Not yet at threshold — skip this cycle, try again next time
			return nil
		}

		// Threshold reached — treat as truly gone, reset counter and trigger recreate
		cm.failureCountMu.Lock()
		delete(cm.failureCount, instance.InstanceID)
		cm.failureCountMu.Unlock()

		cm.logger.Warn("Container does not exist after consecutive failures, preparing to recreate",
			zap.String("instance_id", instance.InstanceID),
			zap.String("container_name", instance.ContainerName),
			zap.Int("failures", count))

		return cm.recreateContainerWithStatus(ctx, instance, containerCreateOptions,
			model.ContainerStatusPending, "Container does not exist, recreating")
	}

	// GetInfo succeeded — reset failure counter
	cm.failureCountMu.Lock()
	delete(cm.failureCount, instance.InstanceID)
	cm.failureCountMu.Unlock()

	// Parse container creation time (RFC3339 format or Docker format)
	containerCreatedAt, err := time.Parse(time.RFC3339, containerInfo.CreatedAt)
	if err != nil {
		// Try parsing with Docker default format if RFC3339 fails
		containerCreatedAt, err = time.Parse("2006-01-02 15:04:05", containerInfo.CreatedAt)
		if err != nil {
			cm.logger.Warn("Failed to parse container creation time",
				zap.String("instance_id", instance.InstanceID),
				zap.String("created_at", containerInfo.CreatedAt),
				zap.Error(err))
			return err
		}
	}
	containerCreatedAtMs := containerCreatedAt.UnixMilli()

	// Not equal to running, check startup timeout, if startup timeout then cleanup container
	// Normalize status to lowercase for comparison (Docker uses "running", K8s uses "Running")
	if strings.ToLower(containerInfo.Status) != "running" {
		// Check startup timeout
		if instance.StartupTimeout > 0 {
			if (currentTime - containerCreatedAtMs) > instance.StartupTimeout {
				// Startup timeout, cleanup container and service, update status
				startupDuration := currentTime - containerCreatedAtMs
				cm.logger.Warn("Container startup timeout, cleaning up resources",
					zap.String("instance_id", instance.InstanceID),
					zap.String("container_status", containerInfo.Status),
					zap.Int64("startup_duration_ms", startupDuration),
					zap.Int64("timeout_at_ms", instance.StartupTimeout))

				return cm.cleanupAndUpdateStatus(ctx, instance,
					fmt.Sprintf("Container startup timeout, startup duration: %d milliseconds, timeout time: %s", startupDuration,
						time.UnixMilli(containerCreatedAtMs).Format(time.RFC3339)))
			}
		}
	}

	// Check if container is ready
	isReady, runInfo, err := containerManager.IsReady(ctx, instance.ContainerName)
	if err != nil {
		cm.logger.Error("Failed to check container ready status",
			zap.String("instance_id", instance.InstanceID),
			zap.String("container_name", instance.ContainerName),
			zap.Error(err))
		return err
	}

	// Handle based on container ready status
	if !isReady {
		// Check startup timeout
		if instance.StartupTimeout > 0 {
			if (currentTime - containerCreatedAtMs) > instance.StartupTimeout {
				// Startup timeout, cleanup container and service, update status
				startupDuration := currentTime - containerCreatedAtMs
				cm.logger.Warn("Container startup timeout, cleaning up resources",
					zap.String("instance_id", instance.InstanceID),
					zap.String("container_status", containerInfo.Status),
					zap.Int64("startup_duration_ms", startupDuration),
					zap.Int64("timeout_at_ms", instance.StartupTimeout))

				return cm.cleanupAndUpdateStatus(ctx, instance,
					fmt.Sprintf("Container startup timeout, startup duration: %d milliseconds, timeout time: %s", startupDuration,
						time.UnixMilli(instance.StartupTimeout).Format(time.RFC3339)))
			}
		}

		// Check running timeout (container started but not ready)
		if strings.ToLower(containerInfo.Status) == "running" {
			if instance.RunningTimeout > 0 && (currentTime-containerCreatedAtMs) > instance.RunningTimeout {
				// Running timeout but not ready, update instance status
				runningDuration := currentTime - containerCreatedAtMs
				cm.logger.Warn("Container running but not ready, running timeout",
					zap.String("instance_id", instance.InstanceID),
					zap.String("container_name", instance.ContainerName),
					zap.Int64("running_duration_ms", runningDuration),
					zap.Int64("timeout_at_ms", instance.RunningTimeout),
					zap.String("run_info", runInfo))

				instance.ContainerIsReady = false
				instance.ContainerStatus = model.ContainerStatusRunTimeoutStop
				instance.ContainerLastMessage = fmt.Sprintf("Container running but not ready, running timeout, running duration: %d milliseconds, status info: %s", runningDuration, runInfo)
				err := cm.instanceRepo.Update(ctx, instance)
				if err != nil {
					return fmt.Errorf("failed to update instance status: %w", err)
				}
			}
		}

		// Instance container status is running, but not ready, update instance status
		if instance.ContainerStatus == model.ContainerStatusRunning {
			instance.ContainerIsReady = false
			instance.ContainerStatus = model.ContainerStatusRunningUnready
			instance.ContainerLastMessage = "Container running but not ready"
			err := cm.instanceRepo.Update(ctx, instance)
			if err != nil {
				return fmt.Errorf("failed to update instance status: %w", err)
			}
		}
		// Container still starting or running but not ready, continue waiting
		cm.logger.Debug("Container not ready, continue waiting",
			zap.String("instance_id", instance.InstanceID),
			zap.String("container_status", containerInfo.Status),
			zap.String("run_info", runInfo))

	} else {
		// Container is ready: check running timeout
		if instance.RunningTimeout > 0 {
			if (currentTime - containerCreatedAtMs) > instance.RunningTimeout {
				// Running timeout, update instance status
				runningDuration := currentTime - containerCreatedAtMs
				message := fmt.Sprintf("Container running timeout, running duration: %d milliseconds, timeout time: %s", runningDuration, time.UnixMilli(instance.RunningTimeout).Format(time.RFC3339))
				cm.logger.Warn("Container running timeout",
					zap.String("instance_id", instance.InstanceID),
					zap.String("container_name", instance.ContainerName),
					zap.Int64("running_duration_ms", runningDuration),
					zap.Int64("timeout_at_ms", instance.RunningTimeout))

				instance.ContainerIsReady = false
				instance.ContainerStatus = model.ContainerStatusRunTimeoutStop
				instance.ContainerLastMessage = message
				err := cm.instanceRepo.Update(ctx, instance)
				if err != nil {
					return fmt.Errorf("failed to update instance status: %w", err)
				}
			}
		}

		// Container running normally and ready
		cm.logger.Debug("Container running normally and ready",
			zap.String("instance_id", instance.InstanceID),
			zap.String("container_name", instance.ContainerName))

		// Ensure instance status is running
		if instance.ContainerStatus != model.ContainerStatusRunning {
			instance.ContainerIsReady = true
			instance.ContainerStatus = model.ContainerStatusRunning
			instance.ContainerLastMessage = "Container running normally and ready"
			err := cm.instanceRepo.Update(ctx, instance)
			if err != nil {
				return fmt.Errorf("failed to update instance status: %w", err)
			}
		}
	}

	return nil
}

// cleanupAndUpdateStatus cleans up container and service, and updates status to startup timeout stop.
// Even if delete fails, DB status is updated — the orphan cleanup task will handle the residual container.
func (cm *ContainerMonitorImpl) cleanupAndUpdateStatus(ctx context.Context, instance *model.McpInstance, message string) error {
	// Use Biz layer to delete container and service
	_, err := biz.GContainerBiz.DeleteContainer(instance)
	if err != nil {
		cm.logger.Warn("Failed to cleanup container resources — orphan cleanup will handle it",
			zap.String("instance_id", instance.InstanceID),
			zap.Error(err))
		// Continue to update DB status regardless of cleanup failure
	}

	instance.ContainerIsReady = false
	instance.ContainerStatus = model.ContainerStatusInitTimeoutStop
	instance.ContainerLastMessage = message
	err = cm.instanceRepo.Update(ctx, instance)
	if err != nil {
		return fmt.Errorf("failed to update instance status: %w", err)
	}
	return nil
}

// recreateContainerWithStatus recreates container and sets specified status
func (cm *ContainerMonitorImpl) recreateContainerWithStatus(ctx context.Context, instance *model.McpInstance,
	options *container.ContainerCreateOptions, containerStatus model.ContainerStatus, message string) error {

	// 1. Delete old resources (using Biz layer)
	_, err := biz.GContainerBiz.DeleteContainer(instance)
	if err != nil {
		cm.logger.Warn("Failed to delete old container resources during recreation",
			zap.String("instance_id", instance.InstanceID),
			zap.Error(err))
		// Continue to create new container
	}

	// 2. Create new resources (using Biz layer)
	// Pass 0 for startupTimeout to disable context timeout for the creation call itself
	err = biz.GContainerBiz.CreateContainer(ctx, options, int32(instance.EnvironmentID), 0)
	if err != nil {
		cm.logger.Error("Failed to create new container",
			zap.String("instance_id", instance.InstanceID),
			zap.Error(err))
		return fmt.Errorf("failed to create new container: %w", err)
	}

	// 3. Update instance information
	instance.ContainerName = options.ContainerName
	instance.ContainerServiceName = options.ServiceName
	instance.ContainerStatus = containerStatus
	instance.ContainerLastMessage = message
	err = cm.instanceRepo.Update(ctx, instance)
	if err != nil {
		cm.logger.Error("Failed to update instance container information",
			zap.String("instance_id", instance.InstanceID),
			zap.String("new_container_name", options.ContainerName),
			zap.Error(err))
		return fmt.Errorf("failed to update instance container information: %w", err)
	}

	cm.logger.Info("Container recreated successfully",
		zap.String("instance_id", instance.InstanceID),
		zap.String("new_container_name", options.ContainerName),
		zap.String("new_service_name", options.ServiceName))

	return nil
}
