package biz

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/kymo-mcp/mcpcan/internal/market/config"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/container"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/kymo-mcp/mcpcan/pkg/k8s"
	"github.com/kymo-mcp/mcpcan/pkg/utils"

	instancepb "github.com/kymo-mcp/mcpcan/api/market/instance"
)

// TaskStatus task status information
// Remove TaskStatus struct, no longer use task management

//go:embed sidecar_hosting_template.yaml
var sidecarHostingTemplate string

//go:embed sidecar_proxy_template.yaml
var sidecarProxyTemplate string

// buildSidecarConfig 将模板中的占位符替换为实际值
func buildSidecarConfig(tmpl string, replacements map[string]string) string {
	cfg := tmpl
	for k, v := range replacements {
		cfg = strings.ReplaceAll(cfg, k, v)
	}
	return cfg
}

// ContainerBiz container data layer
type ContainerBiz struct {
	ctx context.Context
}

var GContainerBiz *ContainerBiz

func init() {
	GContainerBiz = NewContainerBiz(context.Background())
}

// NewContainerBiz create container data processing layer instance
func NewContainerBiz(ctx context.Context) *ContainerBiz {
	return &ContainerBiz{
		ctx: ctx,
	}
}

type ContainerOptions struct {
	// Instance ID
	InstanceID string

	// Container name
	ContainerName string

	// McpServers configuration
	McpServers string

	// Port mapping configuration
	PortMapping map[int]int

	// Initialization script content
	InitScript string

	// Environment variables configuration
	EnvironmentVariables map[string]string

	// Volume mount configuration (supports multiple volumes)
	VolumeMounts []k8s.UnifiedMount

	// Millisecond timestamp, default 0 means no detection and always create, when set the maximum cannot exceed 1 day
	StartupTimeout int64

	// Millisecond timestamp, default 0 means resident service, when set the maximum cannot exceed 1 year (more than 1 year should be set as resident)
	RunningTimeout int64

	// code package download link
	PackageDownloadLink string
}

// Remove no longer used storage and node configuration structs, affinity logic has been moved to Create method

// ContainerCreateResult container creation result
type ContainerCreateResult struct {
	ContainerName string
	ServiceName   string
	ServicePort   int32
	Message       string
}

// ContainerDeleteParams container deletion parameters
type ContainerDeleteParams struct {
	InstanceID string
}

// ContainerDeleteResult container deletion result
type ContainerDeleteResult struct {
	ContainerName string
	ServiceName   string
	Message       string
}

// ContainerStatusParams container status query parameters
type ContainerStatusParams struct {
	InstanceID string
}

// ContainerStatusResult container status query result
type ContainerStatusResult struct {
	ContainerName  string
	ServiceName    string
	ErrorMessage   string
	ContainerReady bool                       // Whether container is ready
	ServiceReady   bool                       // Whether service is ready
	WarningEvents  []container.ContainerEvent // Warning events
}

// CreateContainer create container business logic
func (cd *ContainerBiz) CreateContainer(ctx context.Context, containerCreateOptions *container.ContainerCreateOptions, environmentId int32, startupTimeout int32) error {
	// 9. Set timeout context
	if startupTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(startupTimeout)*time.Second)
		defer cancel()
	}

	entry, err := cd.GetRuntimeEntry(ctx, uint(environmentId))
	if err != nil {
		return errors.New(i18n.FormatWithContext(ctx, i18n.CodeGetRuntimeEntryFailure, err))
	}
	if entry == nil {
		return errors.New(i18n.FormatWithContext(ctx, i18n.CodeContainerRuntimeNotInitialized, "entry is nil"))
	}

	// create container
	containerName, err := entry.GetContainerManager().Create(ctx, *containerCreateOptions)
	if err != nil {
		// Delete container (if container name is not empty)
		if containerName != "" {
			_ = entry.GetContainerManager().Delete(ctx, containerName)
		}
		return errors.New(i18n.FormatWithContext(ctx, i18n.CodeContainerCreateFailure, err))
	}

	// Check runtime type
	if entry.GetRuntimeType() == container.RuntimeDocker {
		// For Docker, we don't create a separate service.
		// The container name acts as the service name (hostname).
		containerCreateOptions.ServiceName = containerCreateOptions.ContainerName
		return nil
	}

	// create service
	_, err = entry.GetServiceManager().Create(ctx, containerCreateOptions.ServiceName, containerCreateOptions.Port, containerCreateOptions.Labels)
	if err != nil {
		// Delete container (if container name is not empty)
		if containerName != "" {
			_ = entry.GetContainerManager().Delete(ctx, containerName)
		}
		return errors.New(i18n.FormatWithContext(ctx, i18n.CodeServiceCreateFailure, err))
	}

	return nil
}

// DeleteContainer delete container business logic
func (cd *ContainerBiz) DeleteContainer(instance *model.McpInstance) (*ContainerDeleteResult, error) {
	if len(instance.ContainerName) <= 0 {
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(cd.ctx, i18n.CodeInstanceContainerNotExists))
	}
	if instance.EnvironmentID <= 0 {
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(cd.ctx, i18n.CodeInstanceEnvironmentIDNotExists))
	}
	entry, err := cd.GetRuntimeEntry(cd.ctx, instance.EnvironmentID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.FormatWithContext(cd.ctx, i18n.CodeGetRuntimeEntryFailure), err)
	}
	if entry == nil {
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(cd.ctx, i18n.CodeContainerRuntimeNotInitialized))
	}

	message := ""
	// 2. Delete container
	if err = entry.GetContainerManager().Delete(cd.ctx, instance.ContainerName); err != nil {
		message += fmt.Sprintf(i18n.FormatWithContext(cd.ctx, i18n.CodeDeleteContainerFailure)+": %v \n", err)
	} else {
		message += i18n.FormatWithContext(cd.ctx, i18n.CodeContainerDeleteSuccess) + " \n"
	}

	// 3. Delete service
	if err = entry.GetServiceManager().Delete(cd.ctx, instance.ContainerServiceName); err != nil {
		message += fmt.Sprintf(i18n.FormatWithContext(cd.ctx, i18n.CodeServiceDeleteFailure)+": %v", err.Error())
	} else {
		message += i18n.FormatWithContext(cd.ctx, i18n.CodeServiceDeleteSuccess) + " \n"
	}

	resp := &ContainerDeleteResult{
		ContainerName: instance.ContainerName,
		ServiceName:   instance.ContainerServiceName,
		Message:       message,
	}
	return resp, nil
}

// GetContainerStatus get detailed container status information, including container exception detection and service probing
func (cd *ContainerBiz) GetContainerStatus(params ContainerStatusParams) (*instancepb.GetStatusResp, error) {
	// 1. Get instance configuration based on instanceID
	instance, err := mysql.McpInstanceRepo.FindByInstanceIDAndAccessType(
		context.Background(),
		params.InstanceID,
		model.AccessTypeHosting, // Only hosting mode needs to query container status
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.FormatWithContext(cd.ctx, i18n.CodeInstanceNotHostingMode), err)
	}
	if len(instance.ContainerName) <= 0 {
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(cd.ctx, i18n.CodeInstanceContainerNotExists))
	}

	if instance.EnvironmentID <= 0 {
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(cd.ctx, i18n.CodeInstanceEnvironmentIDNotExists))
	}

	entry, err := cd.GetRuntimeEntry(cd.ctx, instance.EnvironmentID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.FormatWithContext(cd.ctx, i18n.CodeGetRuntimeEntryFailure), err)
	}
	if entry == nil {
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(cd.ctx, i18n.CodeContainerRuntimeNotInitialized))
	}

	message := ""
	warningEvents := make([]container.ContainerEvent, 0)
	// 3. Check container ready status
	containerReady, runInfo, err := entry.GetContainerManager().IsReady(cd.ctx, instance.ContainerName)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.FormatWithContext(cd.ctx, i18n.CodeContainerReadyCheckFailure), err)
	}
	if !containerReady {

		message += fmt.Sprintf(i18n.FormatWithContext(cd.ctx, i18n.CodeContainerNotReady)+": %s \n", runInfo)
		// 4. Get container warning events
		warningEvents, err = entry.GetContainerManager().GetWarningEvents(cd.ctx, instance.ContainerName)
		if err != nil {
			message += fmt.Sprintf(i18n.FormatWithContext(cd.ctx, i18n.CodeGetContainerWarningEventsFailure)+": %v \n", err)
		}
	}

	// 5. Actively probe whether the service is running normally
	svc, svcErr := entry.GetServiceManager().Get(cd.ctx, instance.ContainerServiceName)
	svcReady := false
	if svcErr == nil {
		// Check if service configuration is normal
		if svc.ClusterIP != "" {
			// For Headless Service, ClusterIP being "None" is also normal
			if svc.ClusterIP == "None" || svc.ClusterIP == "docker-network" {
				// Headless Service or Docker network, check if there is port configuration
				svcReady = len(svc.Ports) > 0
			} else {
				// Normal Service, check ClusterIP and port configuration
				svcReady = len(svc.Ports) > 0
			}
		}
	} else {
		message += fmt.Sprintf(i18n.FormatWithContext(cd.ctx, i18n.CodeServiceStatusAbnormal)+": %v \n", svcErr)
	}

	// 6. Update instance information
	if containerReady && svcReady {
		instance.ContainerStatus = model.ContainerStatusRunning
		instance.ContainerIsReady = true
		instance.ContainerLastMessage = message
	} else {
		instance.ContainerIsReady = false
		instance.ContainerLastMessage = message
	}
	err = mysql.McpInstanceRepo.Update(context.Background(), instance)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.FormatWithContext(cd.ctx, i18n.CodeUpdateInstanceFailure), err)
	}

	events := make([]*instancepb.ContainerEvent, 0, len(warningEvents))
	for _, event := range warningEvents {
		events = append(events, &instancepb.ContainerEvent{
			Type:          event.Type,
			Reason:        event.Reason,
			Message:       event.Message,
			LastTimestamp: event.Timestamp,
		})
	}

	// Use HTTP probe to check service availability (since we moved back to container cross network probing)
	probeResult := utils.ProbePortFromURL(cd.ctx, instance.ContainerServiceURL, 5*time.Second)

	probeHttp := false
	if probeResult.Success {
		probeHttp = true
	} else {
		message += fmt.Sprintf("HTTP probe failed: %s", probeResult.Error)
	}

	resp := &instancepb.GetStatusResp{
		InstanceId:     params.InstanceID,
		Status:         string(instance.Status),
		ContainerName:  instance.ContainerName,
		RuntimeType:    string(entry.GetRuntimeType()),
		ContainerReady: containerReady,
		ServiceReady:   svcReady,
		ProbeHttp:      probeHttp,
		WarningEvents:  events,
		ErrorMessage:   message,
	}

	return resp, nil
}

// generateContainerName generates container name
func (cd *ContainerBiz) generateContainerName(instanceID string) string {
	// Generate container name based on instance ID
	instanceID = instanceID[:8]
	return fmt.Sprintf("mcp-instance-%s-container", instanceID)
}

// generateServiceName generates service name
func (cd *ContainerBiz) generateServiceName(instanceID string) string {
	instanceID = instanceID[:8]
	return fmt.Sprintf("mcp-instance-%s-service", instanceID)
}

type imageParams struct {
	image       string
	port        int32
	command     []string
	commandArgs []string
}

func (cd *ContainerBiz) getMcpHostingImageCfg(instanceID string, imgAddress string, port int32, initScript string, codepkgInstallScript string, mcpServerCfg string) (*imageParams, error) {
	if len(imgAddress) == 0 {
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(cd.ctx, i18n.CodeImageAddressRequired))
	}
	if port == 0 {
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(cd.ctx, i18n.CodePortRequired))
	}
	if len(initScript) == 0 {
		initScript = "echo 'No initialization commands specified'"
	}

	// 整理脚本，去除首尾空格
	codepkgInstallScript = strings.TrimSpace(codepkgInstallScript)
	initScript = strings.TrimSpace(initScript)

	// Build complete startup script
	// Escape single quotes for shell echo command
	mcpServerCfg = strings.ReplaceAll(mcpServerCfg, "'", "'\\''")
	startupScript := fmt.Sprintf(`
# Create working directory
mkdir -p /app/init

# Generate initialization script dynamically
cat > /app/init/startup.sh << 'EOF_STARTUP'
#!/bin/sh
set -e

echo "[$(date)] --- Startup Script Stage 1: Write Config ---"
echo '%s' > /app/mcp-servers.json

echo "[$(date)] --- Startup Script Stage 2: Code Package ---"
%s

echo "[$(date)] --- Startup Script Stage 3: Init Script ---"
# Execute initialization script
%s

echo "[$(date)] --- Startup Script Stage 4: Main Command ---"
echo "[$(date)] Starting main program: mcp-hosting --port=%d"
if [ -f "/usr/local/bin/mcp-hosting" ]; then
    exec mcp-hosting --port=%d --mcp-servers-config /app/mcp-servers.json
else
    echo "Error: mcp-hosting binary not found at /usr/local/bin/mcp-hosting"
    exit 1
fi
EOF_STARTUP

# Set script execution permissions
chmod +x /app/init/startup.sh

# Execute startup command script
exec /app/init/startup.sh
`,
		mcpServerCfg,
		codepkgInstallScript,
		initScript,
		port,
		port)

	imgPms := &imageParams{
		image:       imgAddress,
		port:        port,
		command:     []string{"/bin/sh"},
		commandArgs: []string{"-c", startupScript},
	}

	return imgPms, nil
}

func (cd *ContainerBiz) getMcpHostingImageCfgForSSEAndSteamableHttp(instanceID string, imgAddress string, port int32, initScript string, command string, codepkgInstallScript string) (*imageParams, error) {
	if len(imgAddress) == 0 {
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(cd.ctx, i18n.CodeImageAddressRequired))
	}
	if len(command) == 0 {
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(cd.ctx, i18n.CodeStartupCommandRequired))
	}

	// 整理脚本，去除首尾空格
	codepkgInstallScript = strings.TrimSpace(codepkgInstallScript)
	initScript = strings.TrimSpace(initScript)
	command = strings.TrimSpace(command)

	// Build complete startup script
	startupScript := fmt.Sprintf(`
# Create working directory
mkdir -p /app/init

# Generate initialization script dynamically
cat > /app/init/startup.sh << 'EOF_STARTUP'
#!/bin/sh
set -e

echo "[$(date)] --- Startup Script Stage 1: Code Package ---"
%s

echo "[$(date)] --- Startup Script Stage 2: Init Script ---"
%s

echo "[$(date)] --- Startup Script Stage 3: Main Command ---"
echo "[$(date)] Starting startup command: %s"
%s
EOF_STARTUP
# Set script execution permissions
chmod +x /app/init/startup.sh

# Execute startup command script
exec /app/init/startup.sh
`,
		codepkgInstallScript, initScript, command, command)

	imgPms := &imageParams{
		image:       imgAddress,
		port:        port,
		command:     []string{"/bin/sh"},
		commandArgs: []string{"-c", startupScript},
	}

	return imgPms, nil
}

// ContainerScaleParams container scaling parameters
type ContainerScaleParams struct {
	InstanceID string
	Replicas   int32
}

// ContainerScaleResult container scaling result
type ContainerScaleResult struct {
	Message string
}

// ContainerLogsParams container logs parameters
type ContainerLogsParams struct {
	InstanceID string
	Lines      int64
}

// ContainerRestartResult container restart result
type ContainerRestartResult struct {
	ContainerName string
	Message       string
}

// ScaleContainerToZero scales container replicas to 0
func (cd *ContainerBiz) ScaleContainerToZero(instance *model.McpInstance) (*ContainerScaleResult, error) {
	// 1. Get instance configuration based on instanceID
	instance, err := mysql.McpInstanceRepo.FindByInstanceIDAndAccessType(
		context.Background(),
		instance.InstanceID,
		model.AccessTypeHosting, // Only hosting mode needs container scaling
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.FormatWithContext(cd.ctx, i18n.CodeInstanceNotHostingMode), err)
	}
	if len(instance.ContainerName) <= 0 {
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(cd.ctx, i18n.CodeInstanceContainerNotExists))
	}
	if instance.EnvironmentID <= 0 {
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(cd.ctx, i18n.CodeInstanceEnvironmentIDNotExists))
	}

	entry, err := cd.GetRuntimeEntry(cd.ctx, instance.EnvironmentID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.FormatWithContext(cd.ctx, i18n.CodeGetRuntimeEntryFailure), err)
	}
	if entry == nil {
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(cd.ctx, i18n.CodeContainerRuntimeNotInitialized))
	}

	// Get container manager and service manager
	containerManager := entry.GetContainerManager()

	// Choose scaling strategy based on runtime type
	if instance.ContainerName != "" {
		// Get runtime type
		runtimeType := entry.GetRuntimeType()

		if runtimeType == container.RuntimeKubernetes {
			// Kubernetes: Set replicas to 0
			e1 := containerManager.Scale(cd.ctx, instance.ContainerName, 0)
			if e1 != nil {
				return nil, fmt.Errorf("%s: %w", i18n.FormatWithContext(cd.ctx, i18n.CodeContainerScaledToZero), e1)
			}
		} else {
			// Docker: Delete container
			e2 := containerManager.Delete(cd.ctx, instance.ContainerName)
			if e2 != nil {
				return nil, fmt.Errorf("%s: %w", i18n.FormatWithContext(cd.ctx, i18n.CodeDeleteContainerFailure), e2)
			}
		}
	}

	// Update instance status
	instance.Status = model.InstanceStatusInactive
	instance.ContainerIsReady = false
	instance.ContainerStatus = model.ContainerStatusManualStop
	instance.ContainerLastMessage = i18n.FormatWithContext(cd.ctx, i18n.CodeContainerScaledToZero)
	err = mysql.McpInstanceRepo.Update(cd.ctx, instance)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.FormatWithContext(cd.ctx, i18n.CodeUpdateInstanceFailure), err)
	}

	return &ContainerScaleResult{Message: i18n.FormatWithContext(cd.ctx, i18n.CodeContainerScaledToZero)}, nil
}

// GetContainerLogs get container logs
func (cd *ContainerBiz) GetContainerLogs(ctx context.Context, params ContainerLogsParams) (string, error) {
	instance, err := mysql.McpInstanceRepo.FindByInstanceID(ctx, params.InstanceID)
	if err != nil {
		return "", errors.New(i18n.FormatWithContext(ctx, i18n.CodeInstanceQueryFailure, err))
	}
	if instance == nil {
		return "", fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeInstanceNotExists))
	}
	if instance.AccessType != model.AccessTypeHosting {
		return "", fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeInstanceNotHostingMode))
	}
	if len(instance.ContainerName) <= 0 {
		return "", fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeInstanceContainerNotExists))
	}
	if instance.EnvironmentID <= 0 {
		return "", fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeInstanceEnvironmentIDNotExists))
	}

	entry, err := cd.GetRuntimeEntry(ctx, instance.EnvironmentID)
	if err != nil {
		return "", errors.New(i18n.FormatWithContext(ctx, i18n.CodeGetRuntimeEntryFailure, err))
	}
	if entry == nil {
		return "", fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeContainerRuntimeNotInitialized))
	}

	// Set default number of lines
	lines := params.Lines
	if lines <= 0 {
		lines = 100
	}

	// Get container logs
	logs, err := entry.GetContainerManager().GetLogs(ctx, instance.ContainerName, lines)
	if err != nil {
		return "", errors.New(i18n.FormatWithContext(ctx, i18n.CodeGetContainerLogsFailure, err))
	}

	return logs, nil
}

// RestartContainer container restart business logic
func (cd *ContainerBiz) RestartContainer(ctx context.Context, instance *model.McpInstance) (*ContainerRestartResult, error) {
	entry, err := cd.GetRuntimeEntry(ctx, instance.EnvironmentID)
	if err != nil {
		return nil, errors.New(i18n.FormatWithContext(ctx, i18n.CodeGetRuntimeEntryFailure, err))
	}
	if entry == nil {
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeContainerRuntimeNotInitialized))
	}

	if len(instance.ContainerName) <= 0 {
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeInstanceContainerNotExists))
	}

	// Parse container creation options
	var containerOptions container.ContainerCreateOptions
	if instance.AccessType == model.AccessTypeHosting {
		// Rebuild options to incorporate latest logic (e.g. Sidecar path fixes)
		var sourceCfg struct {
			McpServers string `json:"mcpServers"`
		}
		if len(instance.SourceConfig) > 0 {
			_ = json.Unmarshal(instance.SourceConfig, &sourceCfg)
		}

		var evs map[string]string
		if len(instance.EnvironmentVariables) > 0 {
			_ = json.Unmarshal(instance.EnvironmentVariables, &evs)
		}

		var vms []*instancepb.VolumeMount
		if len(instance.VolumeMounts) > 0 {
			_ = json.Unmarshal(instance.VolumeMounts, &vms)
		}

		newOptions, err := cd.BuildContainerOptions(ctx, instance.InstanceID, instance.McpProtocol,
			sourceCfg.McpServers, instance.PackageID, instance.Port, instance.InitScript, instance.Command, instance.ImgAddr,
			evs, vms, int32(instance.StartupTimeout), int32(instance.RunningTimeout))
		if err != nil {
			log.Printf("[RestartContainer] Warning: failed to rebuild options for instance %s: %v, falling back to stored options", instance.InstanceID, err)
			if len(instance.ContainerCreateOptions) > 0 {
				if e2 := json.Unmarshal(instance.ContainerCreateOptions, &containerOptions); e2 != nil {
					return nil, errors.New(i18n.FormatWithContext(ctx, i18n.CodeParseContainerOptionsFailure, e2))
				}
			} else {
				return nil, errors.New(i18n.FormatWithContext(ctx, i18n.CodeMissingContainerOptions))
			}
		} else {
			containerOptions = *newOptions
		}
	} else {
		if len(instance.ContainerCreateOptions) > 0 {
			if e2 := json.Unmarshal(instance.ContainerCreateOptions, &containerOptions); e2 != nil {
				return nil, errors.New(i18n.FormatWithContext(ctx, i18n.CodeParseContainerOptionsFailure, e2))
			}
		} else {
			return nil, errors.New(i18n.FormatWithContext(ctx, i18n.CodeMissingContainerOptions))
		}
	}

	// Ensure container name is consistent with instance
	containerOptions.ContainerName = instance.ContainerName

	// Call container manager's restart method
	err = entry.GetContainerManager().Restart(ctx, containerOptions)
	if err != nil {
		return nil, errors.New(i18n.FormatWithContext(ctx, i18n.CodeRestartContainerFailure, err))
	}

	if entry.GetRuntimeType() == container.RuntimeKubernetes {
		// Get service
		err = entry.GetServiceManager().Restart(ctx, containerOptions)
		if err != nil {
			return nil, errors.New(i18n.FormatWithContext(ctx, i18n.CodeRestartContainerFailure, err))
		}
	}

	return &ContainerRestartResult{
		ContainerName: instance.ContainerName,
		Message:       i18n.FormatWithContext(ctx, i18n.CodeRestartContainerSuccess),
	}, nil
}

// createDownloadLink creates download link
// 使用固定的服务名（容器网络内的 DNS）和 HTTP 端口构建下载地址，
// 供实例容器在同一 Docker 网络中访问 mcp-market 服务下载资源。
func (cd *ContainerBiz) createDownloadLink(downloadLinkPath string) string {
	host := "host.docker.internal"
	port := config.GlobalConfig.Server.HttpPort
	if port <= 0 {
		port = 8080
	}
	return fmt.Sprintf("http://%s:%d/%s/%s",
		host,
		port,
		strings.TrimPrefix(common.GetMarketRoutePrefix(), "/"),
		strings.TrimPrefix(downloadLinkPath, "/"))
}

// volumeMountFromPb converts pb volume mount to local structure
func (cd *ContainerBiz) volumeMountFromPb(vm *instancepb.VolumeMount) k8s.UnifiedMount {
	unifiedMount := k8s.UnifiedMount{
		Type:       k8s.MountType(vm.Type),
		MountPath:  vm.MountPath,
		ReadOnly:   vm.ReadOnly,
		SubPath:    vm.SubPath,
		NodeName:   vm.NodeName,
		HostPath:   vm.HostPath,
		PVCName:    vm.PvcName,
		VolumeName: vm.VolumeName,
	}
	return unifiedMount
}

// generateCodePkgScript generates code package startup script
func (cd *ContainerBiz) generateCodePkgInstallScript(packageId string) (string, error) {
	codepkgInstallScript := ""
	// Find code package
	codePackage, err := mysql.McpCodePackageRepo.FindByPackageID(cd.ctx, packageId)
	if err != nil {
		return codepkgInstallScript, fmt.Errorf("%s: %w", i18n.FormatWithContext(cd.ctx, i18n.CodeFailedToFindCodePackage), err)
	}
	// ext := codePackage.PackageType

	downloadLinkPath := fmt.Sprintf("/code/download/%s", packageId)
	pkgLink := cd.createDownloadLink(downloadLinkPath)
	if codePackage == nil {
		return codepkgInstallScript, fmt.Errorf("code package is nil")
	}
	// 增加对本地已挂载代码包的兼容性支持，如果目录已存在且非空，则跳过下载
	codepkgInstallScript = fmt.Sprintf(`
if [ -d "/app/codepkg" ] && [ "$(ls -A /app/codepkg 2>/dev/null)" ]; then
    echo "[$(date)] Local code package detected at /app/codepkg. Checking for subfolders..."
else
    echo "[$(date)] No code package found at /app/codepkg. Starting download from: %s"
    mkdir -p /app/codepkg /tmp/download
    cd /tmp/download
    if wget -q -O package.zip "%s" || curl -sL -o package.zip "%s"; then
        echo "[$(date)] Package download completed. Extracting to /app/codepkg..."
        unzip -q -o package.zip -d /app/codepkg
        rm package.zip
    else
        echo "[$(date)] ERROR: Failed to download package from %s"
        exit 1
    fi
fi

# Compatibility Logic: Ensure 'mcp-sys-monitor' directory exists (link it if necessary)
if [ ! -d "/app/codepkg/mcp-sys-monitor" ]; then
    actual_dir=$(ls -d /app/codepkg/* 2>/dev/null | grep -v "mcp-sys-monitor" | head -n 1)
    if [ -n "$actual_dir" ] && [ -d "$actual_dir" ]; then
        echo "[$(date)] Creating compatibility symlink: /app/codepkg/mcp-sys-monitor -> $actual_dir"
        ln -sf "$actual_dir" /app/codepkg/mcp-sys-monitor
    fi
fi

echo "[$(date)] Final /app/codepkg structure:"
ls -F /app/codepkg
cd /app
`, pkgLink, pkgLink, pkgLink, pkgLink)
	return codepkgInstallScript, nil
}

// GetRuntimeEntry gets runtime entry for environment
func (ed *ContainerBiz) GetRuntimeEntry(ctx context.Context, environmentID uint) (*container.Entry, error) {
	// Get environment information by environment ID
	environment, err := GEnvironmentBiz.GetEnvironment(ctx, environmentID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.FormatWithContext(ctx, i18n.CodeGetEnvironmentInfoFailure), err)
	}

	// Create different runtime configurations based on environment type
	switch environment.Environment {
	case model.McpEnvironmentKubernetes:
		// Create Kubernetes container runtime entry
		cfg, err := ed.getKubernetesRuntimeConfig(ctx, environment)
		if err != nil {
			return nil, errors.New(i18n.FormatWithContext(ctx, i18n.CodeGetK8sRuntimeEntryFailure, err))
		}
		// Create Kubernetes container runtime entry
		return container.NewEntry(cfg)
	case model.McpEnvironmentDocker:
		// Create Docker container runtime entry
		cfg, err := ed.getDockerRuntimeConfig(ctx, environment)
		if err != nil {
			return nil, errors.New(i18n.FormatWithContext(ctx, i18n.CodeGetEnvironmentInfoFailure, err))
		}
		return container.NewEntry(cfg)
	default:
		return nil, fmt.Errorf("%s", i18n.FormatWithContext(ctx, i18n.CodeUnsupportedEnvironmentType))
	}
}

// getKubernetesRuntimeConfig gets runtime configuration for Kubernetes environment
func (ed *ContainerBiz) getKubernetesRuntimeConfig(ctx context.Context, environment *model.McpEnvironment) (container.Config, error) {
	// Create Kubernetes container runtime configuration
	cfg := common.SetKubeConfig([]byte(environment.Config))
	if cfg == nil {
		return container.Config{}, fmt.Errorf("kubeconfig is empty")
	}

	return container.Config{
		Runtime:    container.RuntimeKubernetes,
		Namespace:  environment.Namespace,
		Kubeconfig: cfg,
	}, nil
}

// getDockerRuntimeConfig gets runtime configuration for Docker environment
func (ed *ContainerBiz) getDockerRuntimeConfig(ctx context.Context, environment *model.McpEnvironment) (container.Config, error) {
	network := environment.DockerNetwork
	// If DockerNetwork is empty, try to use default user-defined network for development environment
	if network == "" {
		return container.Config{}, fmt.Errorf("docker network is empty")
	}

	return container.Config{
		Runtime: container.RuntimeDocker,
		Docker: container.DockerConfig{
			DockerHost:     environment.DockerHost,
			DockerUseTLS:   environment.DockerUseTLS,
			DockerCAData:   environment.DockerCaData,
			DockerCertData: environment.DockerCertData,
			DockerKeyData:  environment.DockerKeyData,
			Network:        network,
		},
	}, nil
}

// BuildContainerOptions builds container creation options
func (cd *ContainerBiz) BuildContainerOptions(ctx context.Context, instanceID string, mcpProtocol model.McpProtocol, mcpServices string, packageId string, port int32, initScript string, command string, imgAddress string,
	evs map[string]string, vms []*instancepb.VolumeMount, startupTimeout int32, runningTimeout int32) (*container.ContainerCreateOptions, error) {
	var err error
	containerName := cd.generateContainerName(instanceID)
	serviceName := cd.generateServiceName(instanceID)
	// Generate code package download link
	codepkgInstallScript := ""
	if packageId != "" {
		// Generate code package install script
		var e1 error
		codepkgInstallScript, e1 = cd.generateCodePkgInstallScript(packageId)
		if e1 != nil {
			return nil, fmt.Errorf("failed to generate code package install script: %w", e1)
		}
	}

	imgPms := &imageParams{}
	if mcpProtocol == model.McpProtocolSSE || mcpProtocol == model.McpProtocolStreamableHttp {
		// Generate image configuration
		imgPms, err = cd.getMcpHostingImageCfgForSSEAndSteamableHttp(instanceID, imgAddress, port, initScript, command, codepkgInstallScript)
		if err != nil {
			return nil, fmt.Errorf("failed to get mcp hosting image config: %w", err)
		}
	} else {
		// Generate image configuration
		imgPms, err = cd.getMcpHostingImageCfg(instanceID, imgAddress, port, initScript, codepkgInstallScript, mcpServices)
		if err != nil {
			return nil, fmt.Errorf("failed to get mcp hosting image config: %w", err)
		}
	}
	if imgPms.image == "" || len(imgPms.commandArgs) == 0 || imgPms.port == 0 {
		return nil, fmt.Errorf("build container options failed: image or command or port is empty")
	}

	// Set environment variables
	envVars := make(map[string]string)
	envVars["MCP_INSTANCE_ID"] = instanceID
	envVars["MCP_PORT"] = fmt.Sprintf("%d", imgPms.port)
	envVars["NODE_ENV"] = "production"
	for k, v := range evs {
		envVars[k] = v
	}

	// Set volume mount configuration (affinity judgment logic moved to Create method)
	mounts := []k8s.UnifiedMount{}
	if len(vms) > 0 {
		for _, vm := range vms {
			mounts = append(mounts, cd.volumeMountFromPb(vm))
		}
	}

	// Set labels
	labels := make(map[string]string)
	labels["app"] = containerName
	labels["instance"] = instanceID
	labels["managed-by"] = common.SourceServerName
	if startupTimeout > 0 {
		labels["mcp.startup.timeout"] = fmt.Sprintf("%d", startupTimeout)
	}
	if runningTimeout > 0 {
		labels["mcp.running.timeout"] = fmt.Sprintf("%d", runningTimeout)
	}

	// Traefik support labels — 移动到 Sidecar 容器上
	prefix := common.GetGatewayRoutePrefix()
	strippedPrefix := strings.Trim(prefix, "/")
	instancePath := fmt.Sprintf("/%s/%s", strippedPrefix, instanceID)
	routerName := fmt.Sprintf("mcp-inst-%s", instanceID)

	traefikLabels := make(map[string]string)
	sidecarContainerName := containerName + "-sidecar" // Sidecar 容器名，用于 Traefik service
	traefikLabels["traefik.enable"] = "true"
	// 动态添加针对该实例前缀的 StripPrefix 中间件
	stripMiddleware := fmt.Sprintf("%s-strip", routerName)
	traefikLabels[fmt.Sprintf("traefik.http.middlewares.%s.stripprefix.prefixes", stripMiddleware)] = instancePath
	// 设置路由规则及中间件链 (Auth -> Strip)
	traefikLabels[fmt.Sprintf("traefik.http.routers.%s.rule", routerName)] = fmt.Sprintf("PathPrefix(`%s`)", instancePath)
	traefikLabels[fmt.Sprintf("traefik.http.routers.%s.middlewares", routerName)] = fmt.Sprintf("mcp-auth@file,%s@docker", stripMiddleware)
	// 健壮性关键修复：router 显式指向 sidecar 容器的 service，防止 Traefik 将同 router 下多个容器的负载规则任意合并
	traefikLabels[fmt.Sprintf("traefik.http.routers.%s.service", routerName)] = sidecarContainerName
	traefikLabels[fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", sidecarContainerName)] = "80"

	// 默认禁用主容器的 Traefik 直接发现（由 Sidecar 代劳）
	labels["traefik.enable"] = "false"


	// 使用 embed 模板生成 agentgateway sidecar 配置
	// 8. Build container creation options
	containerOptions := container.ContainerCreateOptions{
		ImageName:     imgPms.image,
		ContainerName: containerName,
		ServiceName:   serviceName,
		Port:          imgPms.port,
		Command:       imgPms.command,
		CommandArgs:   imgPms.commandArgs,
		RestartPolicy: "Always",
		Labels:        labels,
		EnvVars:       envVars,
		Mounts:        mounts,
		WorkingDir:    "/app",
		Platform:      os.Getenv("MCP_HOSTING_PLATFORM"),
		Sidecar: &container.SidecarOptions{
			ImageName:     "mcpcan-sidecar:latest",
			ContainerName: sidecarContainerName,
			Port:          80, // mcpcan-sidecar default port
			EnvVars: map[string]string{
				// 因为自研 proxy 支持 websocket 和所有请求透传，所以无需区分 sse，直接代理到 mcpBackend 的根或具体路径
				"MCP_TARGET_URL":   fmt.Sprintf("http://%s:%d", containerName, imgPms.port),
				"MCP_ROUTE_PREFIX": instancePath,
			},
			Labels:        traefikLabels,
			Platform:      os.Getenv("MCP_HOSTING_PLATFORM"),
		},
	}

	// Create Kubernetes container runtime configuration
	return &containerOptions, nil
}

// BuildOpenapiContainerOptions builds openapi container creation options
func (cd *ContainerBiz) BuildOpenapiContainerOptions(ctx context.Context, instanceID string, openapiFileID string, startupTimeout int32, runningTimeout int32, openapiBaseUrl string) (*container.ContainerCreateOptions, error) {
	containerName := cd.generateContainerName(instanceID)
	serviceName := cd.generateServiceName(instanceID)

	// Set environment variables
	envVars := make(map[string]string)
	envVars["MCP_INSTANCE_ID"] = instanceID
	envVars["MCP_PORT"] = fmt.Sprintf("%d", 8080)
	envVars["NODE_ENV"] = "production"

	// Set labels
	labels := make(map[string]string)
	labels["app"] = containerName
	labels["instance"] = instanceID
	labels["managed-by"] = common.SourceServerName
	if startupTimeout > 0 {
		labels["mcp.startup.timeout"] = fmt.Sprintf("%d", startupTimeout)
	}
	if runningTimeout > 0 {
		labels["mcp.running.timeout"] = fmt.Sprintf("%d", runningTimeout)
	}

	// Traefik support labels
	prefix := common.GetGatewayRoutePrefix()
	prefix = strings.Trim(prefix, "/")
	routerName := fmt.Sprintf("mcp-inst-%s", instanceID)

	instancePath := fmt.Sprintf("/%s/%s/", prefix, instanceID)

	labels["traefik.enable"] = "true"
	labels[fmt.Sprintf("traefik.http.routers.%s.rule", routerName)] = fmt.Sprintf("PathPrefix(`%s`)", instancePath)
	labels[fmt.Sprintf("traefik.http.routers.%s.middlewares", routerName)] = "mcp-auth@file"
	labels[fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", routerName)] = "80"

	// 构建下载链接
	downloadLinkPath := fmt.Sprintf("/openapi/download/%s", openapiFileID)
	downloadLink := cd.createDownloadLink(downloadLinkPath)

	startupScript := fmt.Sprintf(`
		# Create working directory
		mkdir -p /app/init

		# =================【新增 Sidecar 静态配置】=================
		cat > /app/agentgateway.yaml << 'EOF_PROXY'
listeners:
  - name: local-ingress
    address: "0.0.0.0:80"        # 向外部暴露 80 端口给 Traefik 请求
routes:
  - id: "local-route"
    backend_id: "local-backend"
    match:
      pathPrefix: "%s"           # 截取流量并重写 SSE Payload
backends:
  - id: "local-backend"
    servers:
      - url: "http://127.0.0.1:8080" # 代理给真实的同容器 MCP 服务
EOF_PROXY
		# ========================================================

		# Generate initialization script dynamically
		cat > /app/init/startup.sh << 'EOF_STARTUP'
#!/bin/bash
set -e

echo "[$(date)] --- Startup Script Stage 1: Sidecar ---"
echo "[$(date)] Starting local AgentGateway Sidecar..."
agentgateway -f /app/agentgateway.yaml &

echo "[$(date)] --- Startup Script Stage 2: App Prep ---"
echo "[$(date)] Downloading openapi-mcp configuration..."
curl -f '%s' -o /app/run.yaml

echo "[$(date)] --- Startup Script Stage 3: Main Command ---"
echo "[$(date)] Starting openapi-mcp: --base-url=%s"
exec /app/openapi-mcp --no-log-truncation --log-file=>(tee debug.log) --extended --http=:8080 --base-url=%s run.yaml
EOF_STARTUP
		# Set script execution permissions
		chmod +x /app/init/startup.sh
		
		# Execute startup command script
		exec /app/init/startup.sh
	`, instancePath, downloadLink, openapiBaseUrl, openapiBaseUrl)

	// 8. Build container creation options
	containerOptions := container.ContainerCreateOptions{
		ImageName:     common.GetOpenapiToMcpImage(),
		ContainerName: containerName,
		ServiceName:   serviceName,
		Port:          80,                     // Sidecar 监听 80 端口，由 Traefik Label 自动发现
		Command:       []string{"bash", "-c"}, // Use bash for process substitution
		CommandArgs:   []string{startupScript},
		RestartPolicy: "Always",
		Labels:        labels,
		EnvVars:       envVars,
		WorkingDir:    "/app",
		Platform:      os.Getenv("MCP_HOSTING_PLATFORM"), // 支持通过环境变量指定平台，默认为空
	}

	return &containerOptions, nil
}

// BuildProxySidecarOptions 为外部 SSE 实例（Proxy/Direct 类型）构建翻译器容器选项
// 该容器本质上是一个独立的 agentgateway，负责将外部 SSE 流量重写为带网关前缀的路径
func (cd *ContainerBiz) BuildProxySidecarOptions(ctx context.Context, instanceID string, remoteURL string) (*container.ContainerCreateOptions, error) {
	containerName := fmt.Sprintf("mcp-ext-%s", instanceID)
	serviceName := fmt.Sprintf("mcp-ext-svc-%s", instanceID)

	// 解析远程 URL
	u, err := url.Parse(remoteURL)
	if err != nil {
		return nil, fmt.Errorf("invalid remote URL: %w", err)
	}

	host := u.Hostname()
	portStr := u.Port()
	if portStr == "" {
		if u.Scheme == "https" {
			portStr = "443"
		} else {
			portStr = "80"
		}
	}
	path := u.Path
	if path == "" {
		path = "/"
	}

	// Set labels
	labels := make(map[string]string)
	labels["app"] = containerName
	labels["instance"] = instanceID
	labels["managed-by"] = common.SourceServerName
	labels["mcp.instance.type"] = "proxy-translator"

	// Traefik support labels
	prefix := common.GetGatewayRoutePrefix()
	strippedPrefix := strings.Trim(prefix, "/")
	instancePath := fmt.Sprintf("/%s/%s", strippedPrefix, instanceID)
	routerName := fmt.Sprintf("mcp-inst-%s", instanceID)

	labels["traefik.enable"] = "true"
	
	// 动态添加针对该实例前缀的 StripPrefix 中间件
	stripMiddleware := fmt.Sprintf("%s-strip", routerName)
	labels[fmt.Sprintf("traefik.http.middlewares.%s.stripprefix.prefixes", stripMiddleware)] = instancePath

	// 增加 Header 重写中间件，确保 Host 头部与容器名一致
	headersMiddlewareName := fmt.Sprintf("mcp-proxy-headers-%s", instanceID)
	labels[fmt.Sprintf("traefik.http.middlewares.%s.headers.customrequestheaders.Host", headersMiddlewareName)] = containerName

	// 设置路由规则及中间件链
	labels[fmt.Sprintf("traefik.http.routers.%s.rule", routerName)] = fmt.Sprintf("HostRegexp(`{host:.+}`) && PathPrefix(`%s`)", instancePath)
	labels[fmt.Sprintf("traefik.http.routers.%s.middlewares", routerName)] = fmt.Sprintf("mcp-auth@file,%s@docker,%s@docker", stripMiddleware, headersMiddlewareName)
	// 健壮性关键修复：router 显式指向本容器自身的 service (containerName)
	labels[fmt.Sprintf("traefik.http.routers.%s.service", routerName)] = containerName
	labels[fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", containerName)] = "80"

	// 使用 embed 模板生成 agentgateway sidecar 配置（Proxy 模式）
	// 模板基于已验证的 proxy-test.yaml 格式（无 protocol/match 字段，使用 sse 拆分字段）
	agentGatewayYAML := buildSidecarConfig(sidecarProxyTemplate, map[string]string{
		"{{INSTANCE_PATH}}": instancePath,
		"{{REMOTE_HOST}}": host,
		"REMOTE_PORT_PLACEHOLDER": portStr,
		"{{REMOTE_PATH}}": path,
		"{{REMOTE_SCHEME}}": u.Scheme,
	})

	containerOptions := container.ContainerCreateOptions{
		ImageName:     "cr.agentgateway.dev/agentgateway:0.11.1",
		ContainerName: containerName,
		ServiceName:   serviceName,
		Port:          80, // 容器内部监听 80
		// 容器内部自带 Entrypoint，此处使用 -f 加载下发的配置文件
		CommandArgs:   []string{"-f", "/ag-config.yaml"},
		RestartPolicy: "Always",
		Labels:        labels,
		EnvVars:       map[string]string{"RUST_LOG": "debug"},
		Platform:      "", 
		ConfigContent: agentGatewayYAML,
	}

	return &containerOptions, nil
}
