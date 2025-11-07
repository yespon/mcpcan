package redis

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
	"go.uber.org/zap"
)

const (
	// InstanceCacheKeyPrefix Redis缓存key前缀
	InstanceCacheKeyPrefix = "mcp_instance_key:"
	// InstanceCacheExpire 默认缓存过期时间（1分钟）
	InstanceCacheExpire = 60 * time.Second
	// CachePenetrateProtect 缓存穿透保护时间
	CachePenetrateProtect = 5 * time.Second
	// MaxRetryAttempts 最大重试次数
	MaxRetryAttempts = 3
	// CleanInterval 缓存清理间隔
	CleanInterval = 30 * time.Second
	// EnabledLocalHotCache controls whether local hot cache is enabled
	EnabledLocalHotCache = false
)

// CacheInstanceInfo MCP实例信息
type CacheInstanceInfo struct {
	Instance *model.McpInstance
}

// McpInstanceCache MCP实例缓存管理器（纯缓存逻辑）
type McpInstanceCache struct {
	// 防止缓存击穿/雪崩的互斥锁
	cacheMutexes   map[string]*sync.Mutex
	cacheMutexesMu sync.Mutex

	// 空值缓存，防止缓存穿透
	nullCache   map[string]time.Time
	nullCacheMu sync.Mutex

	// 清理停止信号
	stopCh chan struct{}
	// 清理工作池
	wg sync.WaitGroup
}

// NewMcpInstanceCache 创建MCP实例缓存管理器
func NewMcpInstanceCache() *McpInstanceCache {
	cache := &McpInstanceCache{
		cacheMutexes: make(map[string]*sync.Mutex),
		nullCache:    make(map[string]time.Time),
		stopCh:       make(chan struct{}),
	}

	// 启动缓存清理协程
	cache.wg.Add(1)
	go cache.cleanCache()

	return cache
}

// GetCacheInstanceInfo 获取实例信息（带缓存）
func (c *McpInstanceCache) GetCacheInstanceInfo(instanceID string, dataLoader func(string) (*CacheInstanceInfo, error)) (*CacheInstanceInfo, error) {
	if instanceID == "" {
		return nil, errors.New("instanceID cannot be empty")
	}

	cacheKey := c.GenerateCacheKey(instanceID)

	// 1. 检查空值缓存，防止缓存穿透
	if c.IsNullCache(cacheKey) {
		return nil, fmt.Errorf("instance not found: %s", instanceID)
	}

	// 2. 检查本地热点缓存，防止Redis雪崩
	if cachedInfo := c.GetLocalHotCache(cacheKey); cachedInfo != nil {
		return cachedInfo, nil
	}

	// 3. 获取缓存互斥锁，防止缓存击穿
	mutex := c.GetCacheMutex(cacheKey)
	mutex.Lock()
	defer mutex.Unlock()

	// 4. 双重检查，避免重复查询
	if cachedInfo := c.GetLocalHotCache(cacheKey); cachedInfo != nil {
		return cachedInfo, nil
	}

	// 5. 查询Redis缓存
	if cachedInstance := c.GetRedisCache(cacheKey); cachedInstance != nil {
		// 转换为CacheInstanceInfo并设置本地热点缓存
		info := &CacheInstanceInfo{
			Instance: cachedInstance,
		}
		c.SetLocalHotCache(cacheKey, info)
		return info, nil
	}

	// 6. 缓存未命中，使用数据加载器获取数据
	info, err := dataLoader(instanceID)
	if err != nil {
		if errors.Is(err, errors.New("instance not found")) {
			// 设置空值缓存，防止缓存穿透
			c.SetNullCache(cacheKey)
		}
		return nil, err
	}

	// 7. 异步更新缓存（防止缓存雪崩，使用随机过期时间）
	go c.asyncUpdateCache(cacheKey, info)

	return info, nil
}

// GenerateCacheKey 生成缓存key（暴露给业务层使用）
func (c *McpInstanceCache) GenerateCacheKey(instanceID string) string {
	return InstanceCacheKeyPrefix + instanceID
}

// GetCacheMutex 获取指定key的互斥锁（暴露给业务层使用）
func (c *McpInstanceCache) GetCacheMutex(key string) *sync.Mutex {
	c.cacheMutexesMu.Lock()
	defer c.cacheMutexesMu.Unlock()

	if mutex, exists := c.cacheMutexes[key]; exists {
		return mutex
	}

	mutex := &sync.Mutex{}
	c.cacheMutexes[key] = mutex
	return mutex
}

// IsNullCache 检查是否是空值缓存（暴露给业务层使用）
func (c *McpInstanceCache) IsNullCache(key string) bool {
	c.nullCacheMu.Lock()
	defer c.nullCacheMu.Unlock()

	if expireTime, exists := c.nullCache[key]; exists {
		if time.Now().Before(expireTime) {
			return true
		}
		// 过期了，删除空值缓存
		delete(c.nullCache, key)
	}
	return false
}

// SetNullCache 设置空值缓存（暴露给业务层使用）
func (c *McpInstanceCache) SetNullCache(key string) {
	c.nullCacheMu.Lock()
	defer c.nullCacheMu.Unlock()

	c.nullCache[key] = time.Now().Add(CachePenetrateProtect)
}

// GetLocalHotCache 获取本地热点缓存（暴露给业务层使用）
func (c *McpInstanceCache) GetLocalHotCache(key string) *CacheInstanceInfo {
	return GetHotCache(key)
}

// SetLocalHotCache 设置本地热点缓存（暴露给业务层使用）
func (c *McpInstanceCache) SetLocalHotCache(key string, info *CacheInstanceInfo) {
	SetHotCache(key, info)
}

// GetRedisCache 获取Redis缓存（返回*model.McpInstance）
func (c *McpInstanceCache) GetRedisCache(key string) *model.McpInstance {
	redisClient := GetClient()
	if redisClient == nil {
		return nil
	}

	cachedData, err := redisClient.Get(key)
	if err != nil || cachedData == nil {
		return nil
	}

	dataStr, ok := cachedData.(string)
	if !ok || dataStr == "" {
		return nil
	}

	var cachedInstance model.McpInstance
	if err := json.Unmarshal([]byte(dataStr), &cachedInstance); err != nil {
		logger.Error("Failed to unmarshal cached instance", zap.Error(err))
		return nil
	}

	return &cachedInstance
}

// SetRedisCacheInstance 设置Redis缓存（只缓存*model.McpInstance）
func (c *McpInstanceCache) SetRedisCacheInstance(key string, instance *model.McpInstance, expire time.Duration) error {
	redisClient := GetClient()
	if redisClient == nil {
		return errors.New("redis client is nil")
	}

	data, err := json.Marshal(instance)
	if err != nil {
		return fmt.Errorf("failed to marshal instance: %w", err)
	}

	// 重试机制
	for i := 0; i < MaxRetryAttempts; i++ {
		if err := redisClient.Set(key, string(data), expire); err == nil {
			return nil
		}
		if i < MaxRetryAttempts-1 {
			time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
		}
	}

	return fmt.Errorf("failed to set redis cache after %d attempts", MaxRetryAttempts)
}

// SetRedisCacheInstanceFromInfo 从CacheInstanceInfo中提取model.McpInstance并缓存（兼容旧版本）
func (c *McpInstanceCache) SetRedisCacheInstanceFromInfo(key string, info *CacheInstanceInfo, expire time.Duration) error {
	if info == nil {
		return errors.New("instance info is nil")
	}

	// 直接使用info中的Instance字段，它已经是*model.McpInstance类型
	if info.Instance != nil {
		return c.SetRedisCacheInstance(key, info.Instance, expire)
	}

	return errors.New("instance info.Instance is nil")
}

// asyncUpdateCache 异步更新缓存（已废弃，使用SetRedisCacheInstance）
func (c *McpInstanceCache) asyncUpdateCache(cacheKey string, info *CacheInstanceInfo) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Cache update panic recovered", zap.Any("error", r))
		}
	}()

	// 添加随机过期时间，防止缓存雪崩（60-90秒）
	randomExpire := InstanceCacheExpire + time.Duration(rand.Intn(30))*time.Second

	// 更新Redis缓存（兼容旧版本）
	if err := c.SetRedisCacheInstanceFromInfo(cacheKey, info, randomExpire); err != nil {
		logger.Error("Failed to set redis cache", zap.Error(err))
	}

	// 设置本地热点缓存
	c.SetLocalHotCache(cacheKey, info)
}

// cleanCache 定期清理缓存
func (c *McpInstanceCache) cleanCache() {
	defer c.wg.Done()

	ticker := time.NewTicker(CleanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.performCacheCleanup()
		case <-c.stopCh:
			return
		}
	}
}

// performCacheCleanup 执行缓存清理
func (c *McpInstanceCache) performCacheCleanup() {
	// 清理互斥锁（避免内存泄漏）
	c.cacheMutexesMu.Lock()
	if len(c.cacheMutexes) > MaxLocalCacheSize*2 {
		newMutexes := make(map[string]*sync.Mutex, MaxLocalCacheSize)
		for key, mutex := range c.cacheMutexes {
			newMutexes[key] = mutex
		}
		c.cacheMutexes = newMutexes
	}
	c.cacheMutexesMu.Unlock()

	// 清理空值缓存
	c.nullCacheMu.Lock()
	now := time.Now()
	for key, expireTime := range c.nullCache {
		if now.After(expireTime) {
			delete(c.nullCache, key)
		}
	}
	c.nullCacheMu.Unlock()
}

// evictLeastUsedItems 清理最少使用的缓存项
func (c *McpInstanceCache) evictLeastUsedItems() {
	// Hot cache eviction is encapsulated in hot_cache.go and controlled by parameter.
}

// ClearCache 清理指定实例的缓存
func (c *McpInstanceCache) ClearCache(instanceID string) {
	if instanceID == "" {
		return
	}

	cacheKey := c.GenerateCacheKey(instanceID)

	// Clear local hot cache (encapsulated in separate file)
	ClearHotCache(cacheKey)

	// 清理互斥锁
	c.cacheMutexesMu.Lock()
	delete(c.cacheMutexes, cacheKey)
	c.cacheMutexesMu.Unlock()

	// 清理空值缓存
	c.nullCacheMu.Lock()
	delete(c.nullCache, cacheKey)
	c.nullCacheMu.Unlock()

	// 清理Redis缓存
	if redisClient := GetClient(); redisClient != nil {
		if err := redisClient.Del(cacheKey); err != nil {
			logger.Error("Failed to delete redis cache", zap.String("key", cacheKey), zap.Error(err))
		}
	}
}

// ClearAllCache 清理所有缓存
func (c *McpInstanceCache) ClearAllCache() {
	// Clear all local hot cache (encapsulated in separate file)
	ClearAllHotCache()

	// 清理互斥锁
	c.cacheMutexesMu.Lock()
	c.cacheMutexes = make(map[string]*sync.Mutex)
	c.cacheMutexesMu.Unlock()

	// 清理空值缓存
	c.nullCacheMu.Lock()
	c.nullCache = make(map[string]time.Time)
	c.nullCacheMu.Unlock()
}

// Stop 停止缓存管理器
func (c *McpInstanceCache) Stop() {
	close(c.stopCh)
	c.wg.Wait()
}

// GetCacheStats 获取缓存统计信息
func (c *McpInstanceCache) GetCacheStats() map[string]interface{} {
	stats := make(map[string]interface{})

	stats["local_cache_size"] = HotCacheSize()

	c.cacheMutexesMu.Lock()
	stats["mutex_count"] = len(c.cacheMutexes)
	c.cacheMutexesMu.Unlock()

	c.nullCacheMu.Lock()
	stats["null_cache_size"] = len(c.nullCache)
	c.nullCacheMu.Unlock()

	return stats
}

// 全局缓存实例
var globalInstanceCache *McpInstanceCache
var cacheOnce sync.Once

// GetMcpInstanceCache 获取全局MCP实例缓存管理器
func GetMcpInstanceCache() *McpInstanceCache {
	cacheOnce.Do(func() {
		globalInstanceCache = NewMcpInstanceCache()
	})
	return globalInstanceCache
}

// GetMcpCacheInstanceInfo 全局函数，获取实例信息（带缓存）
func GetMcpCacheInstanceInfo(instanceID string, dataLoader func(string) (*CacheInstanceInfo, error)) (*CacheInstanceInfo, error) {
	return GetMcpInstanceCache().GetCacheInstanceInfo(instanceID, dataLoader)
}

// InitMcpInstanceCache 初始化MCP实例缓存
func InitMcpInstanceCache() {
	// 这里只是确保全局缓存实例被创建
	// 实际的清理协程会在NewMcpInstanceCache中启动
	_ = GetMcpInstanceCache()

	// Toggle local hot cache according to the feature flag
	if EnabledLocalHotCache {
		EnableHotCache()
	} else {
		DisableHotCache()
	}
}

// ClearInstanceCache 清理指定实例的缓存
func ClearInstanceCache(instanceID string) {
	GetMcpInstanceCache().ClearCache(instanceID)
}

// ClearAllInstanceCache 清理所有实例缓存
func ClearAllInstanceCache() {
	GetMcpInstanceCache().ClearAllCache()
}
