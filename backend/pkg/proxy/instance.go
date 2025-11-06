package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

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

// McpInstanceService MCP实例业务服务
type McpInstanceService struct {
	cache *redis.McpInstanceCache
}

// NewMcpInstanceService 创建MCP实例业务服务
func NewMcpInstanceService() *McpInstanceService {
	return &McpInstanceService{
		cache: redis.GetMcpInstanceCache(),
	}
}

// GetInstanceInfo 获取实例信息（带缓存）
func (s *McpInstanceService) GetInstanceInfo(instanceID string) (*InstanceInfo, error) {
	if instanceID == "" {
		return nil, errors.New("instanceID cannot be empty")
	}

	// 1. 首先尝试从缓存获取
	cacheKey := s.cache.GenerateCacheKey(instanceID)

	// 检查空值缓存，防止缓存穿透
	if s.cache.IsNullCache(cacheKey) {
		return nil, fmt.Errorf("instance not found: %s", instanceID)
	}

	// 检查本地热点缓存，防止Redis雪崩
	if cachedInfo := s.cache.GetLocalHotCache(cacheKey); cachedInfo != nil {
		return s.buildInstanceInfo(cachedInfo.Instance)
	}

	// 2. 获取缓存互斥锁，防止缓存击穿
	mutex := s.cache.GetCacheMutex(cacheKey)
	mutex.Lock()
	defer mutex.Unlock()

	// 3. 双重检查，避免重复查询
	if cachedInfo := s.cache.GetLocalHotCache(cacheKey); cachedInfo != nil {
		return s.buildInstanceInfo(cachedInfo.Instance)
	}

	// 4. 查询Redis缓存（只缓存*model.McpInstance）
	if cachedInstance := s.cache.GetRedisCache(cacheKey); cachedInstance != nil {
		// 从缓存的model.McpInstance构建完整的InstanceInfo
		info, err := s.buildInstanceInfo(cachedInstance)
		if err != nil {
			return nil, err
		}
		// 设置本地热点缓存（proxy.InstanceInfo -> redis.CacheInstanceInfo）
		s.cache.SetLocalHotCache(cacheKey, &redis.CacheInstanceInfo{
			Instance: info.Instance,
		})
		logger.Debug("Redis cache hit", zap.String("instanceID", instanceID))
		return info, nil
	}

	// 5. 缓存未命中，查询数据库
	instance, err := s.getInstanceFromDB(instanceID)
	if err != nil {
		return nil, err
	}

	// 6. 构建InstanceInfo
	info, err := s.buildInstanceInfo(instance)
	if err != nil {
		return nil, err
	}

	// 7. 设置本地热点缓存（proxy.InstanceInfo -> redis.CacheInstanceInfo）
	s.cache.SetLocalHotCache(cacheKey, &redis.CacheInstanceInfo{
		Instance: info.Instance,
	})

	// 8. 异步更新缓存（只缓存*model.McpInstance到Redis，防止缓存雪崩，使用随机过期时间）
	go s.asyncUpdateCache(cacheKey, instance)

	logger.Debug("Database query completed", zap.String("instanceID", instanceID))
	return info, nil
}

// getInstanceFromDB 从数据库获取实例
func (s *McpInstanceService) getInstanceFromDB(instanceID string) (*model.McpInstance, error) {
	return mysql.McpInstanceRepo.FindByInstanceID(context.Background(), instanceID)
}

// buildInstanceInfo 构建InstanceInfo对象
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

	// 解析Tokens字段
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

// asyncUpdateCache 异步更新缓存（防止缓存雪崩，使用随机过期时间）
func (s *McpInstanceService) asyncUpdateCache(cacheKey string, instance *model.McpInstance) {
	defer func() {
		if r := recover(); r != nil {
			logger.Warn("asyncUpdateCache panic", zap.Any("error", r))
		}
	}()

	// 随机过期时间，防止缓存雪崩（5-15分钟）
	expireTime := time.Duration(300+rand.Intn(600)) * time.Second
	s.cache.SetRedisCacheInstance(cacheKey, instance, expireTime)
}
