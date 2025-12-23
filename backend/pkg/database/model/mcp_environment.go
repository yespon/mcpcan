package model

import (
	"time"
)

type McpEnvironmentType string

const (
	McpEnvironmentKubernetes McpEnvironmentType = "kubernetes"
	McpEnvironmentDocker     McpEnvironmentType = "docker"
)

type McpEnvironmentLevel string

const (
	// McpEnvironmentLevelSystem System environment level, editing and deletion not allowed
	McpEnvironmentLevelSystem McpEnvironmentLevel = "system"
	// McpEnvironmentLevelUser User environment level, allows custom editing
	McpEnvironmentLevelUser McpEnvironmentLevel = "user"
)

type McpEnvironment struct {
	ID             uint                `gorm:"primarykey;autoIncrement;comment:Primary Key ID" json:"ID"`
	Name           string              `gorm:"size:100;not null;comment:Environment Name" json:"name"`
	Environment    McpEnvironmentType  `gorm:"size:20;not null;comment:Runtime Environment (kubernetes/docker)" json:"environment"`
	Config         string              `gorm:"type:text;comment:Connection Configuration" json:"config"`
	Namespace      string              `gorm:"size:100;not null;comment:Namespace" json:"namespace"`
	DockerHost     string              `gorm:"size:255;comment:Docker Host" json:"dockerHost"`
	DockerUseTLS   bool                `gorm:"default:false;comment:Enable TLS" json:"dockerUseTLS"`
	DockerCaData   string              `gorm:"type:text;comment:CA Data" json:"dockerCaData"`
	DockerCertData string              `gorm:"type:text;comment:Cert Data" json:"dockerCertData"`
	DockerKeyData  string              `gorm:"type:text;comment:Key Data" json:"dockerKeyData"`
	DockerNetwork  string              `gorm:"size:100;comment:Network Name" json:"dockerNetwork"`
	CreatorID      string              `gorm:"size:100;not null;comment:Creator ID" json:"creatorID"`
	IsDeleted      bool                `gorm:"default:false;comment:Is Deleted" json:"isDeleted"`
	Level          McpEnvironmentLevel `gorm:"size:20;not null;comment:Environment Level: system (no edit/delete), user (custom edit)" json:"level"`
	CreatedAt      time.Time           `gorm:"type:timestamp(3);not null;comment:Created At" json:"createdAt"`
	UpdatedAt      time.Time           `gorm:"type:timestamp(3);not null;comment:Updated At" json:"updatedAt"`
}

// TableName specifies the table name
func (McpEnvironment) TableName() string {
	return "mcpcan_environment"
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
