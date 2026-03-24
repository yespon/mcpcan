package biz

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"

	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"gorm.io/gorm"
)

//go:embed template.json
var EmbeddedTemplatesFS embed.FS

const (
	defaultOpenapiTemplateName      = "MCPCAN OpenAPI"
	internalEntryBaseURLPlaceholder = "{{MCPCAN_INTERNAL_ENTRY_BASE_URL}}"
	legacyOpenapiTemplateNotes      = "MCPCAN OpenAPI template. The {{CURRENT_DOMAIN}} placeholder in openapi_base_url will be replaced with the current service domain during initialization. To invoke this template, set the server Authorization header with a Bearer token (for example: Bearer xxxxxxx, which can be obtained from the browser console of the current user in the MCPCAN system)."
)

// GetEmbeddedTemplateJSON returns embedded template JSON data
func GetEmbeddedTemplateJSON() []byte {
	data, err := fs.ReadFile(EmbeddedTemplatesFS, "template.json")
	if err != nil {
		log.Printf("Failed to read embedded template.json: %v", err)
		return nil
	}
	return data
}

type embeddedTemplateItem struct {
	Name            string `json:"name"`
	Port            int32  `json:"port"`
	InitScript      string `json:"init_script"`
	Command         string `json:"command"`
	StartupTimeout  int32  `json:"startup_timeout"`
	RunningTimeout  int32  `json:"running_timeout"`
	PackageID       string `json:"package_id"`
	PackageName     string `json:"package_name"`
	SourceType      string `json:"source_type"`
	AccessType      string `json:"access_type"`
	McpProtocol     string `json:"mcp_protocol"`
	McpServers      string `json:"mcp_servers,omitempty"`
	ImgAddress      string `json:"img_address"`
	Notes           string `json:"notes,omitempty"`
	ServicePath     string `json:"service_path,omitempty"`
	IconPath        string `json:"icon_path,omitempty"`
	OpenapiFileName string `json:"openapi_file_name,omitempty"`
	OpenapiBaseUrl  string `json:"openapi_base_url,omitempty"`
}

// initMcpTemplateData initializes data using embedded JSON templates
func (a *App) initMcpTemplateData(ctx context.Context) error {
	log.Printf("Starting MCP template data initialization...")

	embeddedTemplateJSON := GetEmbeddedTemplateJSON()
	if len(embeddedTemplateJSON) == 0 {
		return fmt.Errorf("embedded template json is empty")
	}

	var items []embeddedTemplateItem
	if err := json.Unmarshal(embeddedTemplateJSON, &items); err != nil {
		return fmt.Errorf("failed to unmarshal embedded template json: %w", err)
	}
	if len(items) == 0 {
		log.Printf("No template items found in embedded JSON")
		return nil
	}

	log.Printf("Found MCPTemplate %d template items in embedded JSON", len(items))

	createdCount := 0
	skippedCount := 0

	for _, it := range items {
		// Check for duplicates by name
		existing, err := mysql.McpTemplateRepo.FindByName(ctx, it.Name)
		if err == nil && existing != nil {
			if err := refreshEmbeddedTemplateIfNeeded(ctx, existing, it); err != nil {
				return fmt.Errorf("failed to refresh embedded template '%s': %w", it.Name, err)
			}
			log.Printf("Template '%s' already exists, skipping", it.Name)
			skippedCount++
			continue
		}

		log.Printf("Creating template '%s'", it.Name)

		var tpl *model.McpTemplate
		if it.OpenapiFileName != "" {
			tpl, err = buildTemplateFromOpenapi(ctx, it)
			if err != nil {
				log.Printf("Failed to build openapi template '%s': %v, skipping", it.Name, err)
				skippedCount++
				continue
			}
		} else {
			tpl, err = buildTemplateFromCodePackage(ctx, it)
			if err != nil {
				log.Printf("Failed to build code package template '%s': %v, skipping", it.Name, err)
				skippedCount++
				continue
			}
		}
		if len(it.McpServers) > 0 {
			// Parse JSON string and validate format
			mcpServersBytes := []byte(it.McpServers)
			if err := validateMcpServersConfig(json.RawMessage(mcpServersBytes)); err != nil {
				log.Printf("Invalid McpServers config for template '%s': %v", it.Name, err)
				return fmt.Errorf("invalid McpServers config for template '%s': %w", it.Name, err)
			}
			tpl.McpServers = json.RawMessage(mcpServersBytes)
		}

		if err := mysql.McpTemplateRepo.Create(ctx, tpl); err != nil {
			log.Printf("Failed to create template '%s': %v", it.Name, err)
			return fmt.Errorf("failed to create template '%s': %w", it.Name, err)
		}

		log.Printf("Successfully created template '%s' with ID: %d", it.Name, tpl.ID)
		createdCount++
	}

	log.Printf("MCP template data initialization completed. Created: %d, Skipped: %d", createdCount, skippedCount)
	return nil
}

func buildTemplateFromCodePackage(ctx context.Context, it embeddedTemplateItem) (*model.McpTemplate, error) {
	packageID, err := resolveCodePackageID(ctx, it.PackageName, it.PackageID)
	if err != nil {
		return nil, err
	}

	return &model.McpTemplate{
		Name:           it.Name,
		Port:           it.Port,
		InitScript:     it.InitScript,
		Command:        it.Command,
		StartupTimeout: it.StartupTimeout,
		RunningTimeout: it.RunningTimeout,
		PackageID:      packageID,
		AccessType:     convertAccessType(it.AccessType),
		McpProtocol:    convertMcpProtocol(it.McpProtocol),
		Notes:          it.Notes,
		ServicePath:    it.ServicePath,
		IconPath:       it.IconPath,
		ImgAddress:     it.ImgAddress,
		SourceType:     convertSourceType(it.SourceType),
	}, nil
}

func buildTemplateFromOpenapi(ctx context.Context, it embeddedTemplateItem) (*model.McpTemplate, error) {
	openapiPackageID, err := resolveOpenapiPackageID(ctx, it.OpenapiFileName)
	if err != nil {
		return nil, err
	}

	return &model.McpTemplate{
		Name:           it.Name,
		Port:           it.Port,
		InitScript:     it.InitScript,
		Command:        it.Command,
		PackageID:      openapiPackageID,
		AccessType:     convertAccessType(it.AccessType),
		McpProtocol:    convertMcpProtocol(it.McpProtocol),
		Notes:          it.Notes,
		OpenapiBaseUrl: replaceCurrentDomainPlaceholder(it.OpenapiBaseUrl),
		SourceType:     model.SourceTypeOpenapi,
	}, nil
}

// refreshEmbeddedTemplateIfNeeded 在服务启动时自动修复内置模板的关联数据：
// 1. 始终同步 PackageID → 确保模板引用的 openapi_file_id 与数据库一致（兼容旧环境升级）
// 2. 若满足条件，同时更新 OpenapiBaseUrl / Notes
func refreshEmbeddedTemplateIfNeeded(ctx context.Context, existing *model.McpTemplate, it embeddedTemplateItem) error {
	if existing == nil || it.OpenapiFileName == "" {
		return nil
	}

	changed := false

	// 1. 始终同步 PackageID：从 DB 查出当前 openapi_file_id，确保模板引用正确
	// 旧环境因各种原因（如手动操作、重新初始化）可能导致 package_id 与 openapi_file_id 不匹配，此处自动修复
	currentFileID, err := findOpenapiFileIDByName(ctx, it.OpenapiFileName)
	if err == nil && currentFileID != "" && existing.PackageID != currentFileID {
		log.Printf("Syncing template '%s' PackageID: %s -> %s", it.Name, existing.PackageID, currentFileID)
		existing.PackageID = currentFileID
		changed = true
	}

	// 2. 按需更新 OpenapiBaseUrl / Notes（仅旧版本模板）
	if shouldRefreshEmbeddedOpenapiTemplate(existing, it) {
		existing.OpenapiBaseUrl = it.OpenapiBaseUrl
		existing.Notes = it.Notes
		changed = true
	}

	if !changed {
		return nil
	}
	return mysql.McpTemplateRepo.Update(ctx, existing)
}

// shouldRefreshEmbeddedOpenapiTemplate 判断是否需要刷新模板的 OpenapiBaseUrl / Notes
// PackageID 的同步由 refreshEmbeddedTemplateIfNeeded 独立处理，此函数仅关注 URL/Notes 字段
func shouldRefreshEmbeddedOpenapiTemplate(existing *model.McpTemplate, it embeddedTemplateItem) bool {
	if existing == nil || it.OpenapiFileName == "" || it.Name != defaultOpenapiTemplateName || existing.Name != it.Name {
		return false
	}

	if existing.Notes != legacyOpenapiTemplateNotes && existing.Notes != it.Notes {
		return false
	}

	if strings.Contains(existing.OpenapiBaseUrl, internalEntryBaseURLPlaceholder) {
		return false
	}

	return existing.OpenapiBaseUrl == "" ||
		strings.Contains(existing.OpenapiBaseUrl, "{{CURRENT_DOMAIN}}") ||
		isLegacyOpenapiTemplateURL(existing.OpenapiBaseUrl)
}

// isLegacyOpenapiTemplateURL 识别是否为旧版本的 OpenAPI 基础地址（硬编码了 http/https 且指向 market 后缀）
func isLegacyOpenapiTemplateURL(rawURL string) bool {
	trimmed := strings.TrimRight(strings.TrimSpace(rawURL), "/")
	if trimmed == "" {
		return false
	}
	if strings.Contains(trimmed, internalEntryBaseURLPlaceholder) || strings.Contains(trimmed, "mcp-entry-svc") {
		return false
	}
	if !strings.HasSuffix(trimmed, "/api/market") {
		return false
	}
	return strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://")
}

func findOpenapiFileIDByName(ctx context.Context, originalName string) (string, error) {
	var pkg model.McpOpenapiPackage
	err := mysql.GetDB().Model(&model.McpOpenapiPackage{}).
		WithContext(ctx).
		Where("original_name = ? AND is_deleted = false AND base_openapi_file_id = ''", originalName).
		First(&pkg).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("openapi file not found: %s", originalName)
		}
		return "", fmt.Errorf("failed to query openapi file: %w", err)
	}
	return pkg.OpenapiFileID, nil
}

func resolveOpenapiPackageID(ctx context.Context, openapiFileName string) (string, error) {
	openapiFileID, err := findOpenapiFileIDByName(ctx, openapiFileName)
	if err != nil {
		return "", err
	}
	log.Printf("Found MCPTemplate openapi file '%s' with ID: %s", openapiFileName, openapiFileID)
	return openapiFileID, nil
}

func resolveCodePackageID(ctx context.Context, packageName string, fallbackID string) (string, error) {
	if packageName == "" {
		return fallbackID, nil
	}

	codePackage, err := mysql.McpCodePackageRepo.FindByOriginalName(ctx, packageName)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("Package '%s' not found, using original PackageID: %s", packageName, fallbackID)
			return fallbackID, nil
		}
		return "", fmt.Errorf("failed to query code package '%s': %w", packageName, err)
	}

	log.Printf("Found MCPTemplate existing package '%s' with ID: %s", packageName, codePackage.PackageID)
	return codePackage.PackageID, nil
}

func replaceCurrentDomainPlaceholder(rawURL string) string {
	if rawURL == "" {
		return rawURL
	}
	if !strings.Contains(rawURL, "{{CURRENT_DOMAIN}}") {
		return rawURL
	}

	domain := os.Getenv("CURRENT_DOMAIN")
	if domain == "" {
		if ip, err := common.GetPublicIP(); err == nil && ip != "" {
			domain = ip
		}
	}
	if domain == "" {
		log.Printf("CURRENT_DOMAIN is not set and public IP not found, keep placeholder in URL: %s", rawURL)
		return rawURL
	}

	domain = strings.TrimSpace(domain)
	if strings.HasPrefix(domain, "http://") {
		domain = strings.TrimPrefix(domain, "http://")
	} else if strings.HasPrefix(domain, "https://") {
		domain = strings.TrimPrefix(domain, "https://")
	}
	domain = strings.TrimRight(domain, "/")

	return strings.ReplaceAll(rawURL, "{{CURRENT_DOMAIN}}", domain)
}

// validateMcpServersConfig validates the JSON format of McpServers configuration
func validateMcpServersConfig(rawConfig json.RawMessage) error {
	if len(rawConfig) == 0 {
		return nil // Empty config is valid
	}

	// Try to unmarshal to McpServersConfig structure
	var config struct {
		McpServers map[string]json.RawMessage `json:"mcpServers"`
	}

	if err := json.Unmarshal(rawConfig, &config); err != nil {
		return fmt.Errorf("failed to unmarshal McpServers config: %w", err)
	}

	if len(config.McpServers) == 0 {
		return fmt.Errorf("no mcpServers found in config")
	}

	// Validate each server configuration
	for serverName, serverConfig := range config.McpServers {
		var mcpConfig struct {
			Args      []string `json:"args,omitempty"`
			Command   string   `json:"command,omitempty"`
			Type      string   `json:"type,omitempty"`
			Transport string   `json:"transport,omitempty"`
			URL       string   `json:"url,omitempty"`
		}

		if err := json.Unmarshal(serverConfig, &mcpConfig); err != nil {
			return fmt.Errorf("failed to unmarshal server config for '%s': %w", serverName, err)
		}

		// Determine protocol type based on configuration
		protocolType := determineProtocolType(mcpConfig)
		if protocolType == "" {
			return fmt.Errorf("cannot determine protocol type for server '%s'", serverName)
		}

		// Validate required fields based on protocol type
		if err := validateProtocolFields(protocolType, mcpConfig); err != nil {
			return fmt.Errorf("validation failed for server '%s': %w", serverName, err)
		}
	}

	return nil
}

// determineProtocolType determines the protocol type based on configuration
func determineProtocolType(config struct {
	Args      []string `json:"args,omitempty"`
	Command   string   `json:"command,omitempty"`
	Type      string   `json:"type,omitempty"`
	Transport string   `json:"transport,omitempty"`
	URL       string   `json:"url,omitempty"`
}) string {
	// 1. Check type and transport fields first
	if config.Type != "" {
		return config.Type
	}
	if config.Transport != "" {
		return config.Transport
	}

	// 2. Check URL field
	if config.URL != "" {
		if strings.Contains(strings.ToLower(config.URL), string(model.McpProtocolStreamableHttp)) {
			return string(model.McpProtocolStreamableHttp)
		}
		return string(model.McpProtocolSSE)
	}

	// 3. Check command field
	if config.Command != "" {
		return "stdio"
	}

	// Default return empty string indicating unable to determine
	return ""
}

// validateProtocolFields validates protocol fields
func validateProtocolFields(protocolType string, config struct {
	Args      []string `json:"args,omitempty"`
	Command   string   `json:"command,omitempty"`
	Type      string   `json:"type,omitempty"`
	Transport string   `json:"transport,omitempty"`
	URL       string   `json:"url,omitempty"`
}) error {
	switch protocolType {
	case string(model.McpProtocolStreamableHttp):
		if config.URL == "" {
			return fmt.Errorf("%s protocol requires a valid url field", protocolType)
		}
	case string(model.McpProtocolSSE):
		if config.URL == "" {
			return fmt.Errorf("%s protocol requires a valid url field", protocolType)
		}
	case string(model.McpProtocolStdio):
		if config.Command == "" {
			return fmt.Errorf("%s protocol requires a valid command field", protocolType)
		}
	default:
		return fmt.Errorf("unknown protocol type: %s", protocolType)
	}
	return nil
}

func convertAccessType(s string) model.AccessType {
	switch s {
	case string(model.AccessTypeDirect):
		return model.AccessTypeDirect
	case string(model.AccessTypeProxy):
		return model.AccessTypeProxy
	case string(model.AccessTypeHosting):
		return model.AccessTypeHosting
	default:
		return model.AccessTypeProxy
	}
}

func convertMcpProtocol(s string) model.McpProtocol {
	switch s {
	case string(model.McpProtocolSSE):
		return model.McpProtocolSSE
	case string(model.McpProtocolStreamableHttp):
		return model.McpProtocolStreamableHttp
	case string(model.McpProtocolStdio):
		return model.McpProtocolStdio
	default:
		return model.McpProtocolSSE
	}
}

func convertSourceType(s string) model.SourceType {
	switch s {
	case string(model.SourceTypeMarket):
		return model.SourceTypeMarket
	case string(model.SourceTypeTemplate):
		return model.SourceTypeTemplate
	case string(model.SourceTypeCustom):
		return model.SourceTypeCustom
	case string(model.SourceTypeOpenapi):
		return model.SourceTypeOpenapi
	default:
		return model.SourceTypeCustom
	}
}
