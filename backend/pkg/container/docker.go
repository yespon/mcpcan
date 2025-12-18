package container

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/kymo-mcp/mcpcan/pkg/k8s"
)

// DockerRuntime Docker runtime implementation
type DockerRuntime struct {
	networkName string         // Docker network name
	client      *client.Client // Docker client
	config      DockerConfig   // Docker configuration
}

// NewDockerRuntime creates Docker runtime
func NewDockerRuntime(config DockerConfig) *DockerRuntime {
	networkName := config.Network
	if networkName == "" {
		networkName = "bridge" // default network
	}

	cli, err := initDockerClient(config)
	if err != nil {
		// Log warning but do not fail, to maintain backward compatibility with existing logic
		// that relies on os/exec and doesn't strictly require the client yet.
		fmt.Printf("Warning: Failed to initialize Docker client: %v\n", err)
	}

	return &DockerRuntime{
		networkName: networkName,
		client:      cli,
		config:      config,
	}
}

// initDockerClient initializes the Docker client with TLS support and fallback
func initDockerClient(config DockerConfig) (*client.Client, error) {
	ctx := context.Background()
	var cli *client.Client
	var err error
	var httpClient *http.Client

	hostURL := config.DockerHost
	certDir := config.DockerCertPath
	useTLS := config.DockerUseTLS

	// Configure HTTP/TLS Client if needed
	if useTLS {
		options := tls.Config{InsecureSkipVerify: true}

		// Attempt to load certificates if directory is provided
		if certDir != "" {
			certPath := filepath.Join(certDir, "cert.pem")
			keyPath := filepath.Join(certDir, "key.pem")
			caPath := filepath.Join(certDir, "ca.pem")

			cert, errCert := tls.LoadX509KeyPair(certPath, keyPath)
			caCert, errCA := os.ReadFile(caPath)

			if errCert == nil && errCA == nil {
				caCertPool := x509.NewCertPool()
				caCertPool.AppendCertsFromPEM(caCert)
				options.Certificates = []tls.Certificate{cert}
				options.RootCAs = caCertPool
				options.InsecureSkipVerify = false
			}
		} else if config.DockerCertData != "" && config.DockerKeyData != "" && config.DockerCAData != "" {
			// Load from data strings
			cert, errCert := tls.X509KeyPair([]byte(config.DockerCertData), []byte(config.DockerKeyData))
			if errCert == nil {
				caCertPool := x509.NewCertPool()
				if caCertPool.AppendCertsFromPEM([]byte(config.DockerCAData)) {
					options.Certificates = []tls.Certificate{cert}
					options.RootCAs = caCertPool
					options.InsecureSkipVerify = false
				}
			} else {
				fmt.Printf("Warning: Failed to load Docker certificates from data: %v\n", errCert)
			}
		}

		httpClient = &http.Client{
			Transport: &http.Transport{TLSClientConfig: &options},
		}
	}

	// Build client options
	opts := []client.Opt{
		client.WithAPIVersionNegotiation(),
	}

	if hostURL != "" {
		opts = append(opts, client.WithHost(hostURL))
	}
	if httpClient != nil {
		opts = append(opts, client.WithHTTPClient(httpClient))
	}

	// If no specific config, use FromEnv
	if hostURL == "" && !useTLS {
		opts = append(opts, client.FromEnv)
	}

	// Create Client
	cli, err = client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, err
	}

	// Test connection (Ping)
	_, errPing := cli.Ping(ctx)
	if errPing != nil {
		// If connection fails and we were trying a specific host, fallback to local
		if hostURL != "" || useTLS {
			fmt.Printf("⚠️ Connection to Docker failed (%s): %v. Falling back to local socket (client.FromEnv)...\n", hostURL, errPing)
			cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
			if err != nil {
				return nil, fmt.Errorf("failed to create local Docker client fallback: %w", err)
			}
			// Verify fallback
			if _, err := cli.Ping(ctx); err != nil {
				return nil, fmt.Errorf("failed to connect to local Docker fallback: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to connect to Docker: %w", errPing)
		}
	}

	return cli, nil
}

// GetClient returns the Docker client
func (dr *DockerRuntime) GetClient() *client.Client {
	return dr.client
}

// GetContainerManager gets container manager
func (dr *DockerRuntime) GetContainerManager() ContainerManager {
	return &DockerContainerManager{
		networkName: dr.networkName,
		config:      dr.config,
	}
}

// GetServiceManager gets service manager
func (dr *DockerRuntime) GetServiceManager() ServiceManager {
	return &DockerServiceManager{
		networkName: dr.networkName,
	}
}

// GetVolumeManager gets volume manager
func (dr *DockerRuntime) GetVolumeManager() VolumeManager {
	return &DockerVolumeManager{
		client: dr.client,
	}
}

// GetRuntimeType gets runtime type
func (dr *DockerRuntime) GetRuntimeType() ContainerRuntime {
	return RuntimeDocker
}

// DockerContainerManager Docker container manager implementation
type DockerContainerManager struct {
	networkName string
	config      DockerConfig
}

// DockerContainerInfo Docker container information structure (matching docker inspect output)
type DockerContainerInfo struct {
	ID              string                `json:"Id"`
	Name            string                `json:"Name"`
	State           DockerState           `json:"State"`
	NetworkSettings DockerNetworkSettings `json:"NetworkSettings"`
	Config          DockerConfigInfo      `json:"Config"`
	Created         string                `json:"Created"`
}

type DockerState struct {
	Status string `json:"Status"`
}

type DockerNetworkSettings struct {
	Ports    map[string][]interface{}     `json:"Ports"`
	Networks map[string]DockerNetworkInfo `json:"Networks"`
}

type DockerNetworkInfo struct {
	IPAddress string `json:"IPAddress"`
}

type DockerConfigInfo struct {
	Labels map[string]string `json:"Labels"`
	Image  string            `json:"Image"`
}

// DockerPort Docker port information
type DockerPort struct {
	PrivatePort int32  `json:"privatePort"`
	PublicPort  int32  `json:"publicPort"`
	Type        string `json:"type"`
}

// Create creates container
func (dcm *DockerContainerManager) Create(ctx context.Context, options ContainerCreateOptions) (string, error) {
	// Build docker run command
	args := []string{"run", "-d"}

	// Set container name
	if options.ContainerName != "" {
		args = append(args, "--name", options.ContainerName)
	}

	// Set network
	if dcm.networkName != "" {
		args = append(args, "--network", dcm.networkName)
	}

	// Set restart policy
	if options.RestartPolicy != "" {
		// Convert to lowercase to support "Always" etc.
		policy := strings.ToLower(options.RestartPolicy)

		// Map K8s style policies to Docker style if needed
		switch policy {
		case "always":
			policy = "always"
		case "onfailure":
			policy = "on-failure"
		case "never":
			policy = "no"
		}

		// Validate restart policy
		validPolicies := []string{"no", "on-failure", "always", "unless-stopped"}
		isValid := false
		for _, p := range validPolicies {
			if policy == p {
				isValid = true
				break
			}
		}
		if !isValid {
			return "", fmt.Errorf("invalid restart policy: %s", options.RestartPolicy)
		}
		args = append(args, "--restart", policy)
	}

	// Set working directory
	if options.WorkingDir != "" {
		// Ensure working directory is absolute path
		if !strings.HasPrefix(options.WorkingDir, "/") {
			options.WorkingDir = "/" + options.WorkingDir
		}
		args = append(args, "-w", options.WorkingDir)
	}

	// Expose port (no host mapping)
	if options.Port > 0 {
		args = append(args, "--expose", fmt.Sprintf("%d", options.Port))
	}

	// Set environment variables
	for key, value := range options.EnvVars {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	// Set volume mounts
	for _, mount := range options.Mounts {
		switch mount.Type {
		case k8s.MountTypeHostPath:
			// For Docker, treat HostPath as bind mount
			args = append(args, "-v", fmt.Sprintf("%s:%s", mount.HostPath, mount.MountPath))
			if mount.ReadOnly {
				args = append(args, ":ro")
			}
		case k8s.MountTypeVolume:
			// For Docker, treat Volume as volume mount
			args = append(args, "-v", fmt.Sprintf("%s:%s", mount.VolumeName, mount.MountPath))
			if mount.ReadOnly {
				args = append(args, ":ro")
			}
		default:
			return "", fmt.Errorf("unsupported mount type: %s", mount.Type)
		}
	}

	// Set labels
	for key, value := range options.Labels {
		args = append(args, "--label", fmt.Sprintf("%s=%s", key, value))
	}

	// Set health check
	if options.ReadinessProbe != nil {
		if options.ReadinessProbe.HTTPGet != nil {
			healthCmd := fmt.Sprintf("curl -f http://localhost:%d%s || exit 1",
				options.ReadinessProbe.HTTPGet.Port,
				options.ReadinessProbe.HTTPGet.Path)
			args = append(args, "--health-cmd", healthCmd)
			args = append(args, "--health-interval", "30s")
			args = append(args, "--health-timeout", "3s")
			args = append(args, "--health-retries", "3")
		}
	}

	// Set entry point program (overrides image ENTRYPOINT, must be before image name)
	if len(options.Command) > 0 {
		args = append(args, "--entrypoint", options.Command[0])
	}

	// Add image name
	args = append(args, options.ImageName)

	// Add command arguments (overrides image CMD)
	if len(options.CommandArgs) > 0 {
		args = append(args, options.CommandArgs...)
	}

	// Execute docker run command
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Env = os.Environ()
	if dcm.config.DockerHost != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("DOCKER_HOST=%s", dcm.config.DockerHost))
	}
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to create Docker container: %w", err)
	}

	// Parse output to get container ID
	// When image is not found locally, docker run outputs pull progress.
	// The container ID is always the last line of the output.
	outStr := strings.TrimSpace(string(output))
	lines := strings.Split(outStr, "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[len(lines)-1]), nil
	}

	return outStr, nil
}

// Delete deletes container
func (dcm *DockerContainerManager) Delete(ctx context.Context, containerName string) error {
	// Stop container
	stopCmd := exec.CommandContext(ctx, "docker", "stop", containerName)
	stopCmd.Env = os.Environ()
	if dcm.config.DockerHost != "" {
		stopCmd.Env = append(stopCmd.Env, fmt.Sprintf("DOCKER_HOST=%s", dcm.config.DockerHost))
	}
	_ = stopCmd.Run() // ignore stop error, container might already be stopped

	// Delete container
	rmCmd := exec.CommandContext(ctx, "docker", "rm", containerName)
	rmCmd.Env = os.Environ()
	if dcm.config.DockerHost != "" {
		rmCmd.Env = append(rmCmd.Env, fmt.Sprintf("DOCKER_HOST=%s", dcm.config.DockerHost))
	}
	if err := rmCmd.Run(); err != nil {
		return fmt.Errorf("failed to delete Docker container: %w", err)
	}

	return nil
}

// Scale sets container replica count (only applicable to k8s)
func (dcm *DockerContainerManager) Scale(ctx context.Context, containerName string, replicas int32) error {
	// Docker does not support native scaling (except Swarm/Compose), here we just log
	fmt.Printf("Warning: Scale not supported for Docker runtime\n")
	return nil
}

// Restart restarts container
func (dcm *DockerContainerManager) Restart(ctx context.Context, options ContainerCreateOptions) error {
	// Docker restart is simpler, can directly use docker restart
	// But to support updating configuration (like env vars), we delete and recreate
	if err := dcm.Delete(ctx, options.ContainerName); err != nil {
		// Ignore error if container doesn't exist
		fmt.Printf("Warning: failed to delete container %s during restart: %v\n", options.ContainerName, err)
	}

	_, err := dcm.Create(ctx, options)
	return err
}

// GetInfo gets container information
func (dcm *DockerContainerManager) GetInfo(ctx context.Context, containerName string) (*ContainerInfo, error) {
	cmd := exec.CommandContext(ctx, "docker", "inspect", containerName)
	cmd.Env = os.Environ()
	if dcm.config.DockerHost != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("DOCKER_HOST=%s", dcm.config.DockerHost))
	}
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to inspect Docker container: %w", err)
	}

	var containers []DockerContainerInfo
	if err := json.Unmarshal(output, &containers); err != nil {
		return nil, fmt.Errorf("failed to parse Docker inspect output: %w", err)
	}

	if len(containers) == 0 {
		return nil, fmt.Errorf("container not found")
	}

	c := containers[0]

	// Get IP
	ip := ""
	for _, net := range c.NetworkSettings.Networks {
		ip = net.IPAddress
		break // use first network IP
	}

	// Get ports
	var ports []int32
	for k := range c.NetworkSettings.Ports {
		// k is like "80/tcp"
		parts := strings.Split(k, "/")
		if len(parts) > 0 {
			var p int
			fmt.Sscanf(parts[0], "%d", &p)
			if p > 0 {
				ports = append(ports, int32(p))
			}
		}
	}

	return &ContainerInfo{
		Name:      c.Name,
		Status:    c.State.Status, // running, exited, etc.
		IP:        ip,
		Ports:     ports,
		Labels:    c.Config.Labels,
		CreatedAt: c.Created,
	}, nil
}

// IsReady checks if container is ready
func (dcm *DockerContainerManager) IsReady(ctx context.Context, containerName string) (bool, string, error) {
	info, err := dcm.GetInfo(ctx, containerName)
	if err != nil {
		return false, "", err
	}

	if info.Status == "running" {
		return true, "Running", nil
	}

	return false, info.Status, nil
}

// GetEvents gets container events
func (dcm *DockerContainerManager) GetEvents(ctx context.Context, containerName string) ([]ContainerEvent, error) {
	// Docker events are stream based, here we simulate by getting logs or just returning empty
	// For simplicity, we can return recent logs as "events"
	logs, err := dcm.GetLogs(ctx, containerName, 10)
	if err != nil {
		return nil, err
	}

	// Convert logs to events
	var events []ContainerEvent
	lines := strings.Split(logs, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			events = append(events, ContainerEvent{
				Type:      "Normal",
				Reason:    "Log",
				Message:   line,
				Timestamp: time.Now().Unix(),
			})
		}
	}

	return events, nil
}

// GetLogs gets container logs
func (dcm *DockerContainerManager) GetLogs(ctx context.Context, containerName string, lines int64) (string, error) {
	// Build docker logs command
	args := []string{"logs"}

	// Set line limit
	if lines > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", lines))
	}

	// Add container name
	args = append(args, containerName)

	// Execute command
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Env = os.Environ()
	if dcm.config.DockerHost != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("DOCKER_HOST=%s", dcm.config.DockerHost))
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get Docker container logs: %w, output: %s", err, string(output))
	}

	return string(output), nil
}

// GetWarningEvents gets container warning events
func (dcm *DockerContainerManager) GetWarningEvents(ctx context.Context, containerName string) ([]ContainerEvent, error) {
	// Check if container has error status
	info, err := dcm.GetInfo(ctx, containerName)
	if err != nil {
		return nil, err
	}

	var events []ContainerEvent
	if info.Status != "running" {
		events = append(events, ContainerEvent{
			Type:      "Warning",
			Reason:    "ContainerNotRunning",
			Message:   fmt.Sprintf("container status abnormal: %s", info.Status),
			Timestamp: time.Now().Unix(),
		})
	}

	return events, nil
}

// getContainerIP gets container IP address
func (dcm *DockerContainerManager) getContainerIP(ctx context.Context, containerName string) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", "inspect", "--format", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", containerName)
	output, err := cmd.Output()
	return strings.TrimSpace(string(output)), err
}

// DockerServiceManager Docker service manager implementation (Docker doesn't have native service concept, using network aliases to simulate)
type DockerServiceManager struct {
	networkName string
}

// Create creates service (implemented through network aliases in Docker)
func (dsm *DockerServiceManager) Create(ctx context.Context, serviceName string, port int32, selector map[string]string) (*ServiceInfo, error) {
	// Docker uses network alias for service discovery, so no extra resource creation needed
	return &ServiceInfo{
		Name:      serviceName,
		ClusterIP: "docker-network", // Docker network identifier
		Ports:     []int32{port},
		Labels:    selector,
	}, nil
}

// Delete deletes service
func (dsm *DockerServiceManager) Delete(ctx context.Context, serviceName string) error {
	// Docker uses network alias, cleanup is handled when container is deleted
	return nil
}

// Get gets service information
func (dsm *DockerServiceManager) Get(ctx context.Context, serviceName string) (*ServiceInfo, error) {
	// Infer container name from service name
	// 1. Try direct usage (new behavior: ServiceName == ContainerName)
	containerName := serviceName

	// Inspect container to verify existence and get ports
	cmd := exec.CommandContext(ctx, "docker", "inspect", "--format", "{{json .NetworkSettings.Ports}}", containerName)
	output, err := cmd.Output()
	if err != nil {
		// 2. If failed, try legacy conversion (mcp-instance-xxx-service -> mcp-instance-xxx-container)
		if strings.HasSuffix(serviceName, "-service") {
			legacyName := strings.Replace(serviceName, "-service", "-container", 1)
			cmd = exec.CommandContext(ctx, "docker", "inspect", "--format", "{{json .NetworkSettings.Ports}}", legacyName)
			output, err = cmd.Output()
			if err == nil {
				containerName = legacyName
			} else {
				return nil, fmt.Errorf("service (container) not found: %w", err)
			}
		} else {
			return nil, fmt.Errorf("service (container) not found: %w", err)
		}
	}

	var portsMap map[string][]interface{}
	if err := json.Unmarshal(output, &portsMap); err != nil {
		return nil, fmt.Errorf("failed to parse container ports: %w", err)
	}

	var ports []int32
	for k := range portsMap {
		// k is like "80/tcp"
		parts := strings.Split(k, "/")
		if len(parts) > 0 {
			var p int
			fmt.Sscanf(parts[0], "%d", &p)
			if p > 0 {
				ports = append(ports, int32(p))
			}
		}
	}

	return &ServiceInfo{
		Name:      serviceName,
		ClusterIP: "docker-network",
		Ports:     ports,
		Labels:    make(map[string]string),
	}, nil
}

// Restart restarts service
func (dsm *DockerServiceManager) Restart(ctx context.Context, options ContainerCreateOptions) error {
	// Get existing service information
	existingService, err := dsm.Get(ctx, options.ServiceName)
	if err != nil {
		// If service doesn't exist, directly return error
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("service %s does not exist, cannot restart", options.ServiceName)
		}
		return fmt.Errorf("failed to get service information: %w", err)
	}

	// Delete existing service
	if err := dsm.Delete(ctx, options.ServiceName); err != nil {
		return fmt.Errorf("failed to delete existing service: %w", err)
	}

	// Wait for service to be completely deleted
	if err := dsm.waitForServiceDeletion(ctx, options.ServiceName); err != nil {
		return fmt.Errorf("failed to wait for service deletion completion: %w", err)
	}

	// Recreate service (use original port and labels)
	_, err = dsm.Create(ctx, options.ServiceName, options.Port, existingService.Labels)
	if err != nil {
		return fmt.Errorf("failed to recreate service %s: %w", options.ServiceName, err)
	}

	return nil
}

// waitForServiceDeletion waits for service to be completely deleted
func (dsm *DockerServiceManager) waitForServiceDeletion(ctx context.Context, serviceName string) error {
	const (
		maxRetries    = 15              // maximum retry count
		retryInterval = 1 * time.Second // retry interval
	)

	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("waiting for service deletion was cancelled: %w", ctx.Err())
		default:
		}

		// Check if service still exists
		_, err := dsm.Get(ctx, serviceName)
		if err != nil {
			// If get fails and is NotFound error, deletion is successful
			if strings.Contains(err.Error(), "not found") {
				return nil
			}
			// Other errors continue retrying
		} else {
			// Service still exists, continue waiting
		}

		time.Sleep(retryInterval)
	}
	return fmt.Errorf("waiting for service deletion timed out, exceeded %d seconds", maxRetries)
}

// DockerVolumeManager Docker volume manager implementation
type DockerVolumeManager struct {
	client *client.Client
}

// List lists volumes
func (dvm *DockerVolumeManager) List(ctx context.Context) ([]VolumeInfo, error) {
	if dvm.client == nil {
		return nil, fmt.Errorf("docker client is not initialized")
	}
	volumes, err := dvm.client.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []VolumeInfo
	for _, v := range volumes.Volumes {
		result = append(result, VolumeInfo{
			Name:       v.Name,
			Driver:     v.Driver,
			Mountpoint: v.Mountpoint,
			Labels:     v.Labels,
			Options:    v.Options,
			Scope:      v.Scope,
			CreatedAt:  v.CreatedAt,
			Status:     convertUsageData(v.UsageData),
		})
	}
	return result, nil
}

// Create creates a volume
func (dvm *DockerVolumeManager) Create(ctx context.Context, name string, driver string, labels map[string]string, options map[string]string) (VolumeInfo, error) {
	if dvm.client == nil {
		return VolumeInfo{}, fmt.Errorf("docker client is not initialized")
	}

	vol, err := dvm.client.VolumeCreate(ctx, volume.CreateOptions{
		Name:       name,
		Driver:     driver,
		Labels:     labels,
		DriverOpts: options,
	})
	if err != nil {
		return VolumeInfo{}, err
	}

	return VolumeInfo{
		Name:       vol.Name,
		Driver:     vol.Driver,
		Mountpoint: vol.Mountpoint,
		Labels:     vol.Labels,
		Options:    vol.Options,
		Scope:      vol.Scope,
		CreatedAt:  vol.CreatedAt,
		Status:     convertUsageData(vol.UsageData),
	}, nil
}

// Inspect inspects a volume
func (dvm *DockerVolumeManager) Inspect(ctx context.Context, name string) (VolumeInfo, error) {
	if dvm.client == nil {
		return VolumeInfo{}, fmt.Errorf("docker client is not initialized")
	}

	vol, err := dvm.client.VolumeInspect(ctx, name)
	if err != nil {
		return VolumeInfo{}, err
	}

	return VolumeInfo{
		Name:       vol.Name,
		Driver:     vol.Driver,
		Mountpoint: vol.Mountpoint,
		Labels:     vol.Labels,
		Options:    vol.Options,
		Scope:      vol.Scope,
		CreatedAt:  vol.CreatedAt,
		Status:     convertUsageData(vol.UsageData),
	}, nil
}

// Remove removes a volume
func (dvm *DockerVolumeManager) Remove(ctx context.Context, name string) error {
	if dvm.client == nil {
		return fmt.Errorf("docker client is not initialized")
	}

	return dvm.client.VolumeRemove(ctx, name, false)
}

// Prune removes unused volumes
func (dvm *DockerVolumeManager) Prune(ctx context.Context) (int, error) {
	if dvm.client == nil {
		return 0, fmt.Errorf("docker client is not initialized")
	}

	report, err := dvm.client.VolumesPrune(ctx, filters.Args{})
	if err != nil {
		return 0, err
	}

	return len(report.VolumesDeleted), nil
}

func convertUsageData(data *volume.UsageData) map[string]interface{} {
	if data == nil {
		return nil
	}
	return map[string]interface{}{
		"size":     data.Size,
		"refCount": data.RefCount,
	}
}
