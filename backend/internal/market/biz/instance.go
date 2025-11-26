package biz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"github.com/kymo-mcp/mcpcan/pkg/redis"
	"github.com/kymo-mcp/mcpcan/pkg/utils"
	"gorm.io/gorm"

	instancepb "github.com/kymo-mcp/mcpcan/api/market/instance"
)

// InstanceBiz instance data processing layer
type InstanceBiz struct {
	ctx context.Context
}

// GInstanceBiz global instance data processing layer instance
var GInstanceBiz *InstanceBiz

func init() {
	GInstanceBiz = NewInstanceBiz(context.Background())
}

// NewInstanceBiz create instance data processing layer instance
func NewInstanceBiz(ctx context.Context) *InstanceBiz {
	return &InstanceBiz{
		ctx: ctx,
	}
}

// GetInstance get instance information
func (biz *InstanceBiz) GetInstance(instanceID string) (*model.McpInstance, error) {
	return mysql.McpInstanceRepo.FindByInstanceID(biz.ctx, instanceID)
}

// DisableInstance disable instance
func (biz *InstanceBiz) DisableInstance(instanceID string) (string, error) {
	instance, err := biz.GetInstance(instanceID)
	if err != nil {
		return "", err
	}
	msg := "Instance has been disabled"
	if instance.AccessType == model.AccessTypeHosting {
		res, err := GContainerBiz.DeleteContainer(instance)
		if err != nil {
			return "", err
		}
		msg = res.Message
	}
	instance.Status = model.InstanceStatusInactive
	instance.ContainerIsReady = false
	instance.ContainerStatus = model.ContainerStatusManualStop
	instance.ContainerLastMessage = msg
	return msg, mysql.McpInstanceRepo.Update(biz.ctx, instance)
}

func (biz *InstanceBiz) UpdateInstance(instance *model.McpInstance) error {
	if instance == nil {
		return fmt.Errorf("instance is nil")
	}
	err := mysql.McpInstanceRepo.Update(biz.ctx, instance)
	if err != nil {
		return err
	}
	return biz.UpdateInstanceCache(instance.InstanceID, instance)
}

// DeleteInstance delete instance
func (biz *InstanceBiz) DeleteInstance(instanceID string) error {
	// Get instance by access type
	_, err := mysql.McpInstanceRepo.FindByInstanceID(biz.ctx, instanceID)
	if err != nil {
		return err
	}
	return mysql.McpInstanceRepo.Delete(biz.ctx, instanceID)
}

// updateInstance updates instance cache in Redis using CacheInstanceInfo.
// It generates a cache key and stores the instance with a default expiration.
func (biz *InstanceBiz) UpdateInstanceCache(instanceID string, instance *model.McpInstance) error {
	if instance == nil {
		return fmt.Errorf("cache instance info is nil")
	}
	cache := redis.GetMcpInstanceCache()
	key := cache.GenerateCacheKey(instanceID)
	return cache.SetRedisCacheInstance(key, instance, redis.InstanceCacheExpire)
}

// TokenListByInstanceID lists tokens for an instance with optional filters
func (biz *InstanceBiz) TokenListByInstanceID(req *instancepb.TokenListByInstanceIDRequest) (*instancepb.TokenListByInstanceIDResponse, error) {
	if req.InstanceId == "" {
		return nil, fmt.Errorf("missing required field: instanceId")
	}

	rows, err := mysql.McpTokenRepo.ListByInstanceID(biz.ctx, req.InstanceId)
	if err != nil {
		return nil, fmt.Errorf("failed to list tokens: %v", err)
	}

	list := make([]*instancepb.McpToken, 0, len(rows))
	for _, r := range rows {
		if req.Token != "" && req.Token != r.Token {
			continue
		}
		var headers map[string]string
		var usages []string
		_ = json.Unmarshal(r.Headers, &headers)
		_ = json.Unmarshal(r.Usages, &usages)
		if len(req.Usages) > 0 {
			if !usageIntersect(usages, req.Usages) {
				continue
			}
		}
		if headers == nil {
			headers = make(map[string]string)
		}
		list = append(list, &instancepb.McpToken{
			Id:               int64(r.ID),
			InstanceId:       r.InstanceID,
			Token:            r.Token,
			ExpireAt:         r.ExpireAt,
			PublishAt:        r.PublishAt,
			Usages:           usages,
			EnabledTransport: r.EnabledTransport,
			Headers:          headers,
		})
	}

	return &instancepb.TokenListByInstanceIDResponse{List: list}, nil
}

func (biz *InstanceBiz) DeleteTokenByID(ctx context.Context, id int64) error {
	row, err := mysql.McpTokenRepo.FindByID(ctx, uint(id))
	if err != nil {
		return fmt.Errorf("failed to find token by id: %w", err)
	}
	if row == nil {
		return fmt.Errorf("token not found")
	}
	if err := mysql.McpTokenRepo.DeleteByID(ctx, uint(id)); err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}
	redis.GetMcpTokenCache().Clear(row.InstanceID, row.Token)
	return nil
}

func usageIntersect(a []string, b []string) bool {
	if len(a) == 0 || len(b) == 0 {
		return true
	}
	set := make(map[string]struct{}, len(a))
	for _, x := range a {
		set[x] = struct{}{}
	}
	for _, y := range b {
		if _, ok := set[y]; ok {
			return true
		}
	}
	return false
}

// SaveTokensForInstance creates or updates tokens based on incoming id field
func (biz *InstanceBiz) SaveTokensForInstance(ctx context.Context, tokens []*instancepb.McpToken) error {
	if len(tokens) == 0 {
		return nil
	}
	rows := make([]model.McpToken, 0, len(tokens))
	nowMs := time.Now().UnixMilli()
	for _, t := range tokens {
		if t.InstanceId == "" {
			return fmt.Errorf("missing required field: instanceId")
		}
		headersBytes, _ := json.Marshal(t.Headers)
		usagesBytes, _ := json.Marshal(t.Usages)
		publishAt := t.PublishAt
		if publishAt == 0 {
			publishAt = nowMs
		}
		if t.Id == 0 {
			rows = append(rows, model.McpToken{
				InstanceID:       t.InstanceId,
				Token:            t.Token,
				EnabledTransport: t.EnabledTransport,
				Headers:          json.RawMessage(headersBytes),
				Usages:           json.RawMessage(usagesBytes),
				ExpireAt:         t.ExpireAt,
				PublishAt:        publishAt,
			})
			redis.GetMcpTokenCache().Clear(t.InstanceId, t.Token)
			continue
		}
		existing, err := mysql.McpTokenRepo.FindByID(ctx, uint(t.Id))
		if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to find token by id: %v", err)
		}
		existing.InstanceID = t.InstanceId
		existing.Token = t.Token
		existing.EnabledTransport = t.EnabledTransport
		existing.Headers = json.RawMessage(headersBytes)
		existing.Usages = json.RawMessage(usagesBytes)
		existing.ExpireAt = t.ExpireAt
		existing.PublishAt = publishAt
		if err := mysql.McpTokenRepo.Update(ctx, existing); err != nil {
			return fmt.Errorf("failed to update token: %v", err)
		}
		redis.GetMcpTokenCache().Clear(existing.InstanceID, existing.Token)
	}
	if len(rows) > 0 {
		return mysql.McpTokenRepo.CreateBatch(ctx, rows)
	}
	return nil
}

// ListInstance get instance list
func (biz *InstanceBiz) ListInstance(page, pageSize int32, filters map[string]interface{}, sortBy, sortOrder string) (*instancepb.ListResp, error) {
	// Query data
	instances, total, err := mysql.McpInstanceRepo.FindWithPagination(biz.ctx, page, pageSize, filters, sortBy, sortOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to query instance list: %v", err)
	}

	// envIds
	envIds := make([]string, 0, len(instances))
	for _, instance := range instances {
		envIds = append(envIds, fmt.Sprintf("%d", instance.EnvironmentID))
	}
	envNames, err := mysql.McpEnvironmentRepo.FindNamesByIDs(biz.ctx, envIds)
	if err != nil {
		return nil, fmt.Errorf("failed to query environment names: %v", err)
	}

	// Convert to proto response
	instanceInfos := make([]*instancepb.ListResp_InstanceInfo, 0, len(instances))
	for _, instance := range instances {
		instanceInfo := common.ConvertToInstanceInfo(instance)
		if envName, ok := envNames[fmt.Sprintf("%d", instance.EnvironmentID)]; ok {
			instanceInfo.EnvironmentName = envName
		}
		instanceInfos = append(instanceInfos, instanceInfo)
	}

	return &instancepb.ListResp{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		List:     instanceInfos,
	}, nil
}

// CreateInstance create instance
func (biz *InstanceBiz) CreateInstance(instance *model.McpInstance) error {
	if instance.InstanceName == "" {
		return fmt.Errorf("instance name cannot be empty")
	}
	// Check if name already exists
	existingInstance, err := mysql.McpInstanceRepo.FindByName(biz.ctx, instance.InstanceName)
	if err == nil && existingInstance != nil {
		return fmt.Errorf("instance name %s already exists", instance.InstanceName)
	}
	// Update instance cache
	biz.UpdateInstanceCache(instance.InstanceID, instance)
	return mysql.McpInstanceRepo.Create(biz.ctx, instance)
}

// UpdateInstanceForDirect update instance
func (biz *InstanceBiz) UpdateInstanceForDirect(ctx context.Context, req *instancepb.EditRequest, oriInstance *model.McpInstance) (*instancepb.EditResp, error) {
	// Update basic information
	if req.Name != "" {
		oriInstance.InstanceName = req.Name
	}
	if req.Notes != "" {
		oriInstance.Notes = req.Notes
	}
	oriInstance.IconPath = req.IconPath

	// Validate MCP configuration format
	reqMcpResult, err := utils.ValidateMcpConfig([]byte(req.McpServers))
	if err != nil {
		return nil, fmt.Errorf("failed to validate mcp servers: %w", err)
	}
	if !reqMcpResult.IsValid {
		return nil, fmt.Errorf("mcp servers config is invalid: %s", reqMcpResult.ErrorMessage)
	}
	if reqMcpResult.Url == "" {
		return nil, fmt.Errorf("mcp servers config is invalid: url is empty")
	}

	oriMcpResult, err := utils.ValidateMcpConfig([]byte(oriInstance.SourceConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to validate mcp servers: %w", err)
	}
	if !utils.CompareMcpValidationResult(reqMcpResult, oriMcpResult) {
		sourceConfig := json.RawMessage([]byte(req.McpServers))
		oriInstance.SourceConfig = sourceConfig
	}

	// Save to database
	err = mysql.McpInstanceRepo.Update(ctx, oriInstance)
	if err != nil {
		return nil, fmt.Errorf("failed to update instance: %v", err)
	}
	// Update instance cache
	biz.UpdateInstanceCache(oriInstance.InstanceID, oriInstance)

	accessType, err := common.ConvertToProtoAccessType(oriInstance.AccessType)
	if err != nil {
		return nil, fmt.Errorf("failed to convert access type: %w", err)
	}
	mcpProtocol, err := common.ConvertToProtoMcpProtocol(oriInstance.McpProtocol)
	if err != nil {
		return nil, fmt.Errorf("failed to convert mcp protocol: %w", err)
	}

	resp := &instancepb.EditResp{
		InstanceId:  oriInstance.InstanceID,
		Name:        oriInstance.InstanceName,
		AccessType:  accessType,
		McpProtocol: mcpProtocol,
		Status:      string(model.InstanceStatusActive),
	}

	return resp, nil
}

// UpdateInstanceForProxy update instance
func (biz *InstanceBiz) UpdateInstanceForProxy(ctx context.Context, req *instancepb.EditRequest, oriInstance *model.McpInstance) (*instancepb.EditResp, error) {
	// Update basic information
	if req.Name != "" {
		oriInstance.InstanceName = req.Name
	}
	if req.Notes != "" {
		oriInstance.Notes = req.Notes
	}
	oriInstance.IconPath = req.IconPath

	// Validate MCP configuration format
	reqMcpResult, err := utils.ValidateMcpConfig([]byte(req.McpServers))
	if err != nil {
		return nil, fmt.Errorf("failed to validate mcp servers: %w", err)
	}
	if !reqMcpResult.IsValid {
		return nil, fmt.Errorf("mcp servers config is invalid: %s", reqMcpResult.ErrorMessage)
	}
	if reqMcpResult.Url == "" {
		return nil, fmt.Errorf("mcp servers config is invalid: url is empty")
	}

	oriMcpResult, err := utils.ValidateMcpConfig([]byte(oriInstance.SourceConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to validate mcp servers: %w", err)
	}
	if !utils.CompareMcpValidationResult(reqMcpResult, oriMcpResult) {
		sourceConfig := json.RawMessage([]byte(req.McpServers))
		oriInstance.SourceConfig = sourceConfig
		oriInstance.ProxyProtocol = model.McpProtocol(reqMcpResult.ProtocolType)
		oriInstance.PublicProxyPath = biz.CreatePublicProxyPath(oriInstance.InstanceID, oriInstance.ProxyProtocol)
	}

	// Save to database
	err = mysql.McpInstanceRepo.Update(ctx, oriInstance)
	if err != nil {
		return nil, fmt.Errorf("failed to update instance: %v", err)
	}
	// Update instance cache
	biz.UpdateInstanceCache(oriInstance.InstanceID, oriInstance)

	accessType, err := common.ConvertToProtoAccessType(oriInstance.AccessType)
	if err != nil {
		return nil, fmt.Errorf("failed to convert access type: %w", err)
	}
	mcpProtocol, err := common.ConvertToProtoMcpProtocol(oriInstance.McpProtocol)
	if err != nil {
		return nil, fmt.Errorf("failed to convert mcp protocol: %w", err)
	}
	resp := &instancepb.EditResp{
		InstanceId:  oriInstance.InstanceID,
		Name:        oriInstance.InstanceName,
		AccessType:  accessType,
		McpProtocol: mcpProtocol,
		Status:      string(model.InstanceStatusActive),
	}

	return resp, nil
}

func (biz *InstanceBiz) UpdateInstanceForOpenapi(ctx context.Context, req *instancepb.UpdateOpenapiRequest, oriInstance *model.McpInstance) (*instancepb.EditResp, error) {
	containerOptions, err := GContainerBiz.BuildOpenapiContainerOptions(ctx, req.InstanceId, req.ChooseOpenapiFileID, 0, 0, req.OpenapiBaseUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to build container configuration: %v", err)
	}
	// Create new instance record
	containerCreateOptions, err := common.MarshalAndAssignConfig(containerOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal container create containerCreateOptions: %w", err)
	}

	// Delete old container and svc service
	_, err = GContainerBiz.DeleteContainer(oriInstance)
	if err != nil {
		return nil, fmt.Errorf("failed to delete container: %v", err)
	}

	// Update
	oriInstance.InstanceName = req.Name
	oriInstance.Notes = req.Notes
	oriInstance.ContainerCreateOptions = containerCreateOptions
	oriInstance.ContainerStatus = model.ContainerStatusPending
	oriInstance.ContainerIsReady = false
	oriInstance.IconPath = req.IconPath
	oriInstance.OpenapiBaseUrl = req.OpenapiBaseUrl
	if req.ChooseOpenapiFileID != oriInstance.PackageID {
		oriInstance.PackageID = req.ChooseOpenapiFileID
	}

	// Save to database
	err = mysql.McpInstanceRepo.Update(ctx, oriInstance)
	if err != nil {
		return nil, fmt.Errorf("failed to update instance: %v", err)
	}

	// Update instance cache
	biz.UpdateInstanceCache(oriInstance.InstanceID, oriInstance)

	resp := &instancepb.EditResp{
		InstanceId:  oriInstance.InstanceID,
		Name:        oriInstance.InstanceName,
		AccessType:  instancepb.AccessType_HOSTING,
		McpProtocol: instancepb.McpProtocol_STEAMABLE_HTTP,
		Status:      string(model.InstanceStatusActive),
	}
	return resp, nil
}

// UpdateInstanceForHosting updates instance
func (biz *InstanceBiz) UpdateInstanceForHosting(ctx context.Context, req *instancepb.EditRequest, oriInstance *model.McpInstance) (*instancepb.EditResp, error) {
	var err error
	port := req.Port
	instanceID := req.InstanceId
	packageID := req.PackageId
	initScript := req.InitScript
	command := req.Command
	imgAddress := req.ImgAddress
	envs := req.EnvironmentVariables
	vms := req.VolumeMounts
	startupTimeout := req.StartupTimeout
	runningTimeout := req.RunningTimeout
	mcpServers := req.McpServers

	if oriInstance.McpProtocol == model.McpProtocolStdio {
		if len(mcpServers) == 0 {
			return nil, fmt.Errorf("mcp servers config is empty")
		}
		reqMcpResult, err2 := utils.ValidateMcpConfig([]byte(mcpServers))
		if err2 != nil {
			return nil, fmt.Errorf("failed to validate mcp servers: %w", err2)
		}
		if !reqMcpResult.IsValid {
			return nil, fmt.Errorf("mcp servers config is invalid: %s", reqMcpResult.ErrorMessage)
		}
		if !reqMcpResult.HasCommand {
			return nil, fmt.Errorf("mcp servers config is invalid: command is required")
		}
		oriInstance.SourceConfig = json.RawMessage([]byte(mcpServers))
	}

	newContainerCreateOptions, err := GContainerBiz.BuildContainerOptions(ctx, instanceID, oriInstance.McpProtocol, mcpServers, packageID, port, initScript,
		command, imgAddress, envs, vms, startupTimeout, runningTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to build container configuration: %v", err)
	}
	containerCreateOptions, err := common.MarshalAndAssignConfig(newContainerCreateOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal container create containerCreateOptions: %w", err)
	}

	// Delete old container and svc service
	_, err = GContainerBiz.DeleteContainer(oriInstance)
	if err != nil {
		return nil, fmt.Errorf("failed to delete container: %v", err)
	}

	toEnvs, err := common.MarshalAndAssignConfig(envs)
	if err != nil {
		return nil, fmt.Errorf("failed to convert environment variables: %w", err)
	}
	toVms, err := common.MarshalAndAssignConfig(vms)
	if err != nil {
		return nil, fmt.Errorf("failed to convert volume mounts: %w", err)
	}

	// Create target configuration
	var ProxyProtocol model.McpProtocol
	publicProxyPath := ""
	containerURL := ""
	switch oriInstance.McpProtocol {
	case model.McpProtocolStdio:
		ProxyProtocol = model.McpProtocolStreamableHttp
		publicProxyPath = biz.CreatePublicProxyPath(instanceID, oriInstance.McpProtocol)
		containerURL = fmt.Sprintf("http://%s:%d/%s", newContainerCreateOptions.ServiceName, newContainerCreateOptions.Port, "mcp")
	case model.McpProtocolSSE, model.McpProtocolStreamableHttp:
		ProxyProtocol = oriInstance.McpProtocol
		publicProxyPath = biz.CreatePublicProxyPath(instanceID, oriInstance.McpProtocol)
		containerURL = fmt.Sprintf("http://%s:%d%s", newContainerCreateOptions.ServiceName, newContainerCreateOptions.Port, req.ServicePath)
	default:
		return nil, fmt.Errorf("unsupported mcp protocol: %v", oriInstance.McpProtocol)
	}

	// Update
	oriInstance.InstanceName = req.Name
	oriInstance.Notes = req.Notes
	oriInstance.Port = int32(port)
	oriInstance.InitScript = initScript
	oriInstance.Command = command
	oriInstance.ServicePath = req.ServicePath
	oriInstance.SourceConfig = json.RawMessage([]byte(mcpServers))
	oriInstance.ImgAddr = imgAddress
	oriInstance.EnvironmentVariables = toEnvs
	oriInstance.VolumeMounts = toVms
	oriInstance.StartupTimeout = int64(startupTimeout)
	oriInstance.RunningTimeout = int64(runningTimeout)
	oriInstance.ContainerCreateOptions = containerCreateOptions
	oriInstance.ContainerStatus = model.ContainerStatusPending
	oriInstance.ContainerServiceURL = containerURL
	oriInstance.ContainerIsReady = false
	oriInstance.PublicProxyPath = publicProxyPath
	oriInstance.ProxyProtocol = ProxyProtocol
	oriInstance.IconPath = req.IconPath

	// Save to database
	err = mysql.McpInstanceRepo.Update(ctx, oriInstance)
	if err != nil {
		return nil, fmt.Errorf("failed to update instance: %v", err)
	}
	// Update instance cache
	biz.UpdateInstanceCache(oriInstance.InstanceID, oriInstance)

	accessType, err := common.ConvertToProtoAccessType(oriInstance.AccessType)
	if err != nil {
		return nil, fmt.Errorf("failed to convert access type: %w", err)
	}
	mcpProtocol, err := common.ConvertToProtoMcpProtocol(oriInstance.McpProtocol)
	if err != nil {
		return nil, fmt.Errorf("failed to convert mcp protocol: %w", err)
	}

	resp := &instancepb.EditResp{
		InstanceId:  oriInstance.InstanceID,
		Name:        oriInstance.InstanceName,
		AccessType:  accessType,
		McpProtocol: mcpProtocol,
		Status:      string(model.InstanceStatusActive),
	}
	return resp, nil
}

// GetInstancesByEnvironmentID gets instance list by environment ID
func (biz *InstanceBiz) GetInstancesByEnvironmentID(ctx context.Context, environmentID uint) ([]*model.McpInstance, error) {
	return mysql.McpInstanceRepo.FindByEnvironmentID(ctx, environmentID)
}

// CreatePublicProxyPath creates public proxy configuration
func (biz *InstanceBiz) CreatePublicProxyPath(instanceID string, mcpProtocol model.McpProtocol) string {
	addr := ""
	if mcpProtocol == model.McpProtocolSSE {
		addr = fmt.Sprintf("/%s/%s/%s", strings.Trim(common.GetGatewayRoutePrefix(), "/"), instanceID, mcpProtocol.String())
	} else if mcpProtocol == model.McpProtocolStreamableHttp {
		addr = fmt.Sprintf("/%s/%s/%s", strings.Trim(common.GetGatewayRoutePrefix(), "/"), instanceID, "mcp")
	} else {
		addr = fmt.Sprintf("/%s/%s", strings.Trim(common.GetGatewayRoutePrefix(), "/"), instanceID)
	}
	return addr
}
