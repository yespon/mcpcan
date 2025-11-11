package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
	"github.com/kymo-mcp/mcpcan/pkg/redis"
	"go.uber.org/zap"
)

type InstanceInfo struct {
	InstanceID   string
	AccessType   model.AccessType
	McpProtocol  model.McpProtocol
	Instance     *model.McpInstance
	McpConfig    *model.McpConfig
	EnabledToken bool
	Tokens       []*model.McpToken
}

// McpInstanceService MCP instance business service
type McpInstanceService struct {
	cache *redis.McpInstanceCache
}

// NewMcpInstanceService create MCP instance business service
func NewMcpInstanceService() *McpInstanceService {
	return &McpInstanceService{
		cache: redis.GetMcpInstanceCache(),
	}
}

// GetInstanceInfo get instance info (with cache)
func (s *McpInstanceService) GetInstanceInfo(instanceID string) (*InstanceInfo, error) {
	if instanceID == "" {
		return nil, errors.New("instanceID cannot be empty")
	}

	// 1. First try to get from cache
	cacheKey := s.cache.GenerateCacheKey(instanceID)

	// Check null cache to prevent cache penetration
	if s.cache.IsNullCache(cacheKey) {
		return nil, fmt.Errorf("instance not found: %s", instanceID)
	}

	// Check local hot cache to prevent Redis avalanche
	if cachedInfo := s.cache.GetLocalHotCache(cacheKey); cachedInfo != nil {
		return s.buildInstanceInfo(cachedInfo.Instance)
	}

	// 2. Get cache mutex to prevent cache breakdown
	mutex := s.cache.GetCacheMutex(cacheKey)
	mutex.Lock()
	defer mutex.Unlock()

	// 3. Double check to avoid duplicate queries
	if cachedInfo := s.cache.GetLocalHotCache(cacheKey); cachedInfo != nil {
		return s.buildInstanceInfo(cachedInfo.Instance)
	}

	// 4. Query Redis cache (only cache *model.McpInstance)
	if cachedInstance := s.cache.GetRedisCache(cacheKey); cachedInstance != nil {
		// Build complete InstanceInfo from cached model.McpInstance
		info, err := s.buildInstanceInfo(cachedInstance)
		if err != nil {
			return nil, err
		}
		// Set local hot cache (proxy.InstanceInfo -> redis.CacheInstanceInfo)
		s.cache.SetLocalHotCache(cacheKey, &redis.CacheInstanceInfo{
			Instance: info.Instance,
		})
		logger.Debug("Redis cache hit", zap.String("instanceID", instanceID))
		return info, nil
	}

	// 5. Cache miss, query database
	instance, err := s.getInstanceFromDB(instanceID)
	if err != nil {
		return nil, err
	}

	// 6. Build InstanceInfo
	info, err := s.buildInstanceInfo(instance)
	if err != nil {
		return nil, err
	}

	// 7. Set local hot cache (proxy.InstanceInfo -> redis.CacheInstanceInfo)
	s.cache.SetLocalHotCache(cacheKey, &redis.CacheInstanceInfo{
		Instance: info.Instance,
	})

	// 8. Async update cache (only cache *model.McpInstance to Redis to prevent cache avalanche, use random expiration time)
	go s.asyncUpdateCache(cacheKey, instance)

	logger.Debug("Database query completed", zap.String("instanceID", instanceID))
	return info, nil
}

// getInstanceFromDB get instance from database
func (s *McpInstanceService) getInstanceFromDB(instanceID string) (*model.McpInstance, error) {
	return mysql.McpInstanceRepo.FindByInstanceID(context.Background(), instanceID)
}

// buildInstanceInfo build InstanceInfo object
func (s *McpInstanceService) buildInstanceInfo(instance *model.McpInstance) (*InstanceInfo, error) {
	// Ensure instance is active
	if instance.Status != model.InstanceStatusActive {
		return nil, fmt.Errorf("instance is not active: %s", instance.InstanceID)
	}
	// Check if protocol is stdio
	if instance.ProxyProtocol == model.McpProtocolStdio {
		return nil, fmt.Errorf("stdio protocol is not supported")
	}

	mcpConfig := &model.McpConfig{}
	var err error

	switch instance.AccessType {
	case model.AccessTypeProxy:
		_, _, mcpConfig, err = instance.GetSourceConfig()
		if err != nil {
			return nil, err
		}
	case model.AccessTypeHosting:
		mcpConfig = &model.McpConfig{
			Type:      instance.ProxyProtocol.String(),
			Transport: instance.ProxyProtocol.String(),
			URL:       instance.ContainerServiceURL,
		}
	case model.AccessTypeDirect:
		return nil, fmt.Errorf("stdio protocol is not supported")
	default:
		return nil, fmt.Errorf("unknown access type: %s", instance.AccessType)
	}

	// Parse Tokens field
	var tokens []*model.McpToken
	if instance.Tokens != nil {
		if err := json.Unmarshal(instance.Tokens, &tokens); err != nil {
			logger.Warn("Failed to unmarshal tokens", zap.Error(err))
			tokens = []*model.McpToken{}
		}
	}

	return &InstanceInfo{
		InstanceID:   instance.InstanceID,
		AccessType:   instance.AccessType,
		McpProtocol:  instance.ProxyProtocol,
		Instance:     instance,
		McpConfig:    mcpConfig,
		EnabledToken: instance.EnabledToken,
		Tokens:       tokens,
	}, nil
}

// asyncUpdateCache async update cache (prevent cache avalanche, use random expiration time)
func (s *McpInstanceService) asyncUpdateCache(cacheKey string, instance *model.McpInstance) {
	defer func() {
		if r := recover(); r != nil {
			logger.Warn("asyncUpdateCache panic", zap.Any("error", r))
		}
	}()

	// Random expiration time to prevent cache avalanche (5-15 minutes)
	s.cache.SetRedisCacheInstance(cacheKey, instance, redis.InstanceCacheExpire)
}
