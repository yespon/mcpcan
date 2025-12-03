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

// DockerEnvironmentConfig Docker 环境配置结构
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
	// McpEnvironmentLevelSystem 系统环境等级，不允许编辑删除
	McpEnvironmentLevelSystem McpEnvironmentLevel = "system"
	// McpEnvironmentLevelUser 用户环境等级，允许自定义编辑
	McpEnvironmentLevelUser McpEnvironmentLevel = "user"
)

type McpEnvironment struct {
	ID          uint                `gorm:"primarykey;autoIncrement;comment:主键ID" json:"ID"`
	Name        string              `gorm:"size:100;not null;comment:环境名称" json:"name"`
	Environment McpEnvironmentType  `gorm:"size:20;not null;comment:运行环境 (kubernetes/docker)" json:"environment"`
	Config      string              `gorm:"type:text;comment:连接配置" json:"config"`
	Namespace   string              `gorm:"size:100;not null;comment:命名空间" json:"namespace"`
	CreatorID   string              `gorm:"size:100;not null;comment:创建人ID" json:"creatorID"`
	IsDeleted   bool                `gorm:"default:false;comment:是否删除" json:"isDeleted"`
	Level       McpEnvironmentLevel `gorm:"size:20;not null;comment:环境等级: system 不允许编辑删除, user 允许自定义编辑" json:"level"`
	CreatedAt   time.Time           `gorm:"type:timestamp(3);not null;comment:创建时间" json:"createdAt"`
	UpdatedAt   time.Time           `gorm:"type:timestamp(3);not null;comment:更新时间" json:"updatedAt"`
}

// TableName 指定表名
func (McpEnvironment) TableName() string {
	return "mcpcan_environment"
}

// GetConfig 解析配置字符串为JSON对象
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

// SetConfig 设置配置对象为JSON字符串
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

// IsDeleted 检查环境是否已被删除
func (m *McpEnvironment) IsDeletedRecord() bool {
	return m.IsDeleted
}

// SetCreatedAt 设置创建时间为当前时间
func (m *McpEnvironment) SetCreatedAt() {
	m.CreatedAt = time.Now()
}

// SetUpdatedAt 设置更新时间为当前时间
func (m *McpEnvironment) SetUpdatedAt() {
	m.UpdatedAt = time.Now()
}

// SetDeleted 设置删除状态
func (m *McpEnvironment) SetDeleted() {
	m.IsDeleted = true
	m.UpdatedAt = time.Now()
}

// ClearDeleted 清除删除状态（用于恢复）
func (m *McpEnvironment) ClearDeleted() {
	m.IsDeleted = false
	m.UpdatedAt = time.Now()
}

// PrepareForCreate 准备创建记录（设置创建和更新时间）
func (m *McpEnvironment) PrepareForCreate() {
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	m.IsDeleted = false
}

// PrepareForUpdate 准备更新记录（设置更新时间）
func (m *McpEnvironment) PrepareForUpdate() {
	m.UpdatedAt = time.Now()
}

// PrepareForDelete 准备删除记录（设置删除状态）
func (m *McpEnvironment) PrepareForDelete() {
	m.SetDeleted()
}

// ValidateForCreate 验证创建环境的必要字段
func (m *McpEnvironment) ValidateForCreate() error {
	if m.Name == "" {
		return fmt.Errorf("environment name is required")
	}
	if m.Environment == "" {
		return fmt.Errorf("environment type is required")
	}
	// k8s环境需要校验namespace，docker环境不需要
	if m.Environment == McpEnvironmentKubernetes && m.Namespace == "" {
		return fmt.Errorf("namespace is required for kubernetes environment")
	}

	// 验证环境类型
	if m.Environment != McpEnvironmentKubernetes && m.Environment != McpEnvironmentDocker {
		return fmt.Errorf("invalid environment type: %s", m.Environment)
	}

	return nil
}

// ValidateForUpdate 验证更新环境的必要字段
func (m *McpEnvironment) ValidateForUpdate() error {
	if m.ID == 0 {
		return fmt.Errorf("environment ID is required for update")
	}

	return m.ValidateForCreate()
}

// GetConfigValue 获取配置中的特定值
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

// SetConfigValue 设置配置中的特定值
func (m *McpEnvironment) SetConfigValue(key string, value interface{}) error {
	config, err := m.GetConfig()
	if err != nil {
		return err
	}

	config[key] = value
	return m.SetConfig(config)
}

// GetKubernetesConfig 获取Kubernetes配置（如果环境类型是kubernetes）
func (m *McpEnvironment) GetKubernetesConfig() (map[string]interface{}, error) {
	if m.Environment != McpEnvironmentKubernetes {
		return nil, fmt.Errorf("environment type is not kubernetes")
	}

	return m.GetConfig()
}

// GetDockerConfig 获取Docker配置（如果环境类型是docker）
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

// SetDockerEnvConfig 设置Docker配置
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

// Clone 创建环境的副本
func (m *McpEnvironment) Clone() *McpEnvironment {
	return &McpEnvironment{
		ID:          0, // 新副本不包含ID
		Name:        m.Name + "_copy",
		Environment: m.Environment,
		Config:      m.Config,
		Namespace:   m.Namespace,
		CreatedAt:   time.Time{},
		UpdatedAt:   time.Time{},
		IsDeleted:   false,
	}
}
