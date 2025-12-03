package model

import (
	"fmt"
	"time"
)

// OpenapiFileType openai文档类型
type OpenapiFileType string

func (p OpenapiFileType) String() string {
	return string(p)
}

const (
	OpenapiFileTypeUnknown OpenapiFileType = "unknown"
	OpenapiFileTypeJson    OpenapiFileType = "json"
	OpenapiFileTypeYaml    OpenapiFileType = "yaml"
)

type McpOpenapiPackage struct {
	ID                uint            `gorm:"primarykey;autoIncrement;comment:主键ID" json:"ID"`
	OpenapiFileID     string          `gorm:"size:100;not null;unique;comment:openapi文档ID" json:"openapiFileID"`
	OpenapiFileType   OpenapiFileType `gorm:"size:10;not null;comment:openai文档类型 (json/yaml)" json:"openapiFileType"`
	OpenapiFilePath   string          `gorm:"size:500;not null;comment:存储路径" json:"openapiFilePath"`
	OriginalName      string          `gorm:"size:255;comment:原始文件名" json:"originalName"`
	FileSize          int64           `gorm:"comment:文件大小(字节)" json:"fileSize"`
	BaseOpenapiFileID string          `gorm:"comment:基于哪个 openapi 文档" json:"baseOpenapiFileID"`
	IsDeleted         bool            `gorm:"default:false;comment:是否删除" json:"isDeleted"`
	CreatedAt         time.Time       `gorm:"type:timestamp(3);not null;comment:创建时间" json:"createdAt"`
	UpdatedAt         time.Time       `gorm:"type:timestamp(3);not null;comment:更新时间" json:"updatedAt"`
}

// TableName 指定表名
func (McpOpenapiPackage) TableName() string {
	return "mcpcan_openapi_package"
}

// IsDeletedRecord 检查OpenAPI文档是否已被删除
func (o *McpOpenapiPackage) IsDeletedRecord() bool {
	return o.IsDeleted
}

// SetCreatedAt 设置创建时间为当前时间
func (o *McpOpenapiPackage) SetCreatedAt() {
	o.CreatedAt = time.Now()
}

// SetUpdatedAt 设置更新时间为当前时间
func (o *McpOpenapiPackage) SetUpdatedAt() {
	o.UpdatedAt = time.Now()
}

// SetDeleted 设置删除状态
func (o *McpOpenapiPackage) SetDeleted() {
	o.IsDeleted = true
	o.UpdatedAt = time.Now()
}

// ClearDeleted 清除删除状态（用于恢复）
func (o *McpOpenapiPackage) ClearDeleted() {
	o.IsDeleted = false
	o.UpdatedAt = time.Now()
}

// PrepareForCreate 准备创建记录（设置创建和更新时间）
func (o *McpOpenapiPackage) PrepareForCreate() {
	now := time.Now()
	o.CreatedAt = now
	o.UpdatedAt = now
	o.IsDeleted = false
}

// PrepareForUpdate 准备更新记录（设置更新时间）
func (o *McpOpenapiPackage) PrepareForUpdate() {
	o.UpdatedAt = time.Now()
}

// PrepareForDelete 准备删除记录（设置删除状态）
func (o *McpOpenapiPackage) PrepareForDelete() {
	o.SetDeleted()
}

// ValidateForCreate 验证创建OpenAPI文档的必要字段
func (o *McpOpenapiPackage) ValidateForCreate() error {
	if o.OpenapiFileID == "" {
		return fmt.Errorf("openapi file ID is required")
	}
	if o.OpenapiFileType == "" {
		return fmt.Errorf("openapi file type is required")
	}
	if o.OpenapiFilePath == "" {
		return fmt.Errorf("openapi file path is required")
	}

	// 验证文件类型
	if o.OpenapiFileType != OpenapiFileTypeJson && o.OpenapiFileType != OpenapiFileTypeYaml {
		return fmt.Errorf("invalid openapi file type: %s", o.OpenapiFileType)
	}

	return nil
}

// ValidateForUpdate 验证更新OpenAPI文档的必要字段
func (o *McpOpenapiPackage) ValidateForUpdate() error {
	if o.ID == 0 {
		return fmt.Errorf("openapi package ID is required for update")
	}

	return o.ValidateForCreate()
}
