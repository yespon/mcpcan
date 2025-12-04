package model

import (
	"encoding/json"
	"fmt"
	"time"
)

type McpEnvironmentType string

const (
	McpEnvironmentKubernetes McpEnvironmentType = "kubernetes"
	McpEnvironmentDocker     McpEnvironmentType = "docker"
)

// DockerEnvironmentConfig Docker environment configuration structure
type DockerEnvironmentConfig struct {
	Host     string `json:"host"`     // Docker host (tcp://... or unix://...)
	UseTLS   bool   `json:"useTLS"`   // Enable TLS
	CaData   string `json:"caData"`   // CA certificate data
	CertData string `json:"certData"` // Certificate data
	KeyData  string `json:"keyData"`  // Private key data
	Network  string `json:"network"`  // Network name
}

type McpEnvironmentLevel string

const (
	// McpEnvironmentLevelSystem System environment level, editing and deletion not allowed
	McpEnvironmentLevelSystem McpEnvironmentLevel = "system"
	// McpEnvironmentLevelUser User environment level, allows custom editing
	McpEnvironmentLevelUser McpEnvironmentLevel = "user"
)

type McpEnvironment struct {
	ID          uint                `gorm:"primarykey;autoIncrement;comment:Primary Key ID" json:"ID"`
	Name        string              `gorm:"size:100;not null;comment:Environment Name" json:"name"`
	Environment McpEnvironmentType  `gorm:"size:20;not null;comment:Runtime Environment (kubernetes/docker)" json:"environment"`
	Config      string              `gorm:"type:text;comment:Connection Configuration" json:"config"`
	Namespace   string              `gorm:"size:100;not null;comment:Namespace" json:"namespace"`
	CreatorID   string              `gorm:"size:100;not null;comment:Creator ID" json:"creatorID"`
	IsDeleted   bool                `gorm:"default:false;comment:Is Deleted" json:"isDeleted"`
	Level       McpEnvironmentLevel `gorm:"size:20;not null;comment:Environment Level: system (no edit/delete), user (custom edit)" json:"level"`
	CreatedAt   time.Time           `gorm:"type:timestamp(3);not null;comment:Created At" json:"createdAt"`
	UpdatedAt   time.Time           `gorm:"type:timestamp(3);not null;comment:Updated At" json:"updatedAt"`
}

// TableName specifies the table name
func (McpEnvironment) TableName() string {
	return "mcpcan_environment"
}

// GetConfig parses the configuration string into a JSON object
func (m *McpEnvironment) GetConfig() (map[string]interface{}, error) {
	if m.Config == "" {
		return make(map[string]interface{}), nil
	}

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(m.Config), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

// SetConfig sets the configuration object as a JSON string
func (m *McpEnvironment) SetConfig(config map[string]interface{}) error {
	if config == nil {
		m.Config = ""
		return nil
	}

	configBytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	m.Config = string(configBytes)
	return nil
}

// IsDeletedRecord checks if the environment has been deleted
func (m *McpEnvironment) IsDeletedRecord() bool {
	return m.IsDeleted
}

// SetCreatedAt sets the creation time to the current time
func (m *McpEnvironment) SetCreatedAt() {
	m.CreatedAt = time.Now()
}

// SetUpdatedAt sets the update time to the current time
func (m *McpEnvironment) SetUpdatedAt() {
	m.UpdatedAt = time.Now()
}

// SetDeleted sets the deleted status
func (m *McpEnvironment) SetDeleted() {
	m.IsDeleted = true
	m.UpdatedAt = time.Now()
}

// ClearDeleted clears the deleted status (for restoration)
func (m *McpEnvironment) ClearDeleted() {
	m.IsDeleted = false
	m.UpdatedAt = time.Now()
}

// PrepareForCreate prepares the record for creation (sets created and updated times)
func (m *McpEnvironment) PrepareForCreate() {
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	m.IsDeleted = false
}

// PrepareForUpdate prepares the record for update (sets updated time)
func (m *McpEnvironment) PrepareForUpdate() {
	m.UpdatedAt = time.Now()
}

// PrepareForDelete prepares the record for deletion (sets deleted status)
func (m *McpEnvironment) PrepareForDelete() {
	m.SetDeleted()
}

// ValidateForCreate validates required fields for creating an environment
func (m *McpEnvironment) ValidateForCreate() error {
	if m.Name == "" {
		return fmt.Errorf("environment name is required")
	}
	if m.Environment == "" {
		return fmt.Errorf("environment type is required")
	}
	// k8s environment requires namespace, docker environment does not
	if m.Environment == McpEnvironmentKubernetes && m.Namespace == "" {
		return fmt.Errorf("namespace is required for kubernetes environment")
	}

	// Validate environment type
	if m.Environment != McpEnvironmentKubernetes && m.Environment != McpEnvironmentDocker {
		return fmt.Errorf("invalid environment type: %s", m.Environment)
	}

	return nil
}

// ValidateForUpdate validates required fields for updating an environment
func (m *McpEnvironment) ValidateForUpdate() error {
	if m.ID == 0 {
		return fmt.Errorf("environment ID is required for update")
	}

	return m.ValidateForCreate()
}

// GetConfigValue retrieves a specific value from the configuration
func (m *McpEnvironment) GetConfigValue(key string) (interface{}, error) {
	config, err := m.GetConfig()
	if err != nil {
		return nil, err
	}

	value, exists := config[key]
	if !exists {
		return nil, fmt.Errorf("config key '%s' not found", key)
	}

	return value, nil
}

// SetConfigValue sets a specific value in the configuration
func (m *McpEnvironment) SetConfigValue(key string, value interface{}) error {
	config, err := m.GetConfig()
	if err != nil {
		return err
	}

	config[key] = value
	return m.SetConfig(config)
}

// GetKubernetesConfig retrieves Kubernetes configuration (if environment type is kubernetes)
func (m *McpEnvironment) GetKubernetesConfig() (map[string]interface{}, error) {
	if m.Environment != McpEnvironmentKubernetes {
		return nil, fmt.Errorf("environment type is not kubernetes")
	}

	return m.GetConfig()
}

// GetDockerConfig retrieves Docker configuration (if environment type is docker)
func (m *McpEnvironment) GetDockerConfig() (*DockerEnvironmentConfig, error) {
	if m.Environment != McpEnvironmentDocker {
		return nil, fmt.Errorf("environment type is not docker")
	}

	if m.Config == "" {
		return &DockerEnvironmentConfig{}, nil
	}

	var config DockerEnvironmentConfig
	if err := json.Unmarshal([]byte(m.Config), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal docker config: %w", err)
	}

	return &config, nil
}

// SetDockerEnvConfig sets the Docker configuration
func (m *McpEnvironment) SetDockerEnvConfig(config DockerEnvironmentConfig) error {
	if m.Environment != McpEnvironmentDocker {
		return fmt.Errorf("environment type is not docker")
	}

	configBytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal docker config: %w", err)
	}

	m.Config = string(configBytes)
	return nil
}

// Clone creates a copy of the environment
func (m *McpEnvironment) Clone() *McpEnvironment {
	return &McpEnvironment{
		ID:          0, // New copy does not contain ID
		Name:        m.Name + "_copy",
		Environment: m.Environment,
		Config:      m.Config,
		Namespace:   m.Namespace,
		CreatedAt:   time.Time{},
		UpdatedAt:   time.Time{},
		IsDeleted:   false,
	}
}
