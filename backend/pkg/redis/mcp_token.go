package redis

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
)

const (
	TokenCacheKeyPrefix = "mcp_token_key:"
	TokenCacheExpire    = 60 * time.Second
	TokenNullTTL        = 5 * time.Second
)

type McpTokenCache struct {
	mutexes   map[string]*sync.Mutex
	mutexesMu sync.Mutex
	nulls     map[string]time.Time
	nullsMu   sync.Mutex
}

func NewMcpTokenCache() *McpTokenCache {
	return &McpTokenCache{
		mutexes: make(map[string]*sync.Mutex),
		nulls:   make(map[string]time.Time),
	}
}

func (c *McpTokenCache) GenerateCacheKey(instanceID, token string) string {
	return TokenCacheKeyPrefix + instanceID + ":" + token
}

func (c *McpTokenCache) GetMutex(key string) *sync.Mutex {
	c.mutexesMu.Lock()
	defer c.mutexesMu.Unlock()
	if m, ok := c.mutexes[key]; ok {
		return m
	}
	m := &sync.Mutex{}
	c.mutexes[key] = m
	return m
}

func (c *McpTokenCache) IsNull(key string) bool {
	c.nullsMu.Lock()
	defer c.nullsMu.Unlock()
	if exp, ok := c.nulls[key]; ok {
		if time.Now().Before(exp) {
			return true
		}
		delete(c.nulls, key)
	}
	return false
}

func (c *McpTokenCache) SetNull(key string) {
	c.nullsMu.Lock()
	c.nulls[key] = time.Now().Add(TokenNullTTL)
	c.nullsMu.Unlock()
}

func (c *McpTokenCache) GetRedis(key string) *model.McpToken {
	client := GetClient()
	if client == nil {
		return nil
	}
	v, err := client.Get(key)
	if err != nil || v == nil {
		return nil
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return nil
	}
	var mt model.McpToken
	if json.Unmarshal([]byte(s), &mt) != nil {
		return nil
	}
	return &mt
}

func (c *McpTokenCache) SetRedis(key string, token *model.McpToken, expire time.Duration) error {
	client := GetClient()
	if client == nil {
		return errors.New("redis client is nil")
	}
	b, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("marshal token: %w", err)
	}
	return client.Set(key, string(b), expire)
}

func (c *McpTokenCache) Clear(instanceID, token string) {
	key := c.GenerateCacheKey(instanceID, token)
	c.mutexesMu.Lock()
	delete(c.mutexes, key)
	c.mutexesMu.Unlock()
	c.nullsMu.Lock()
	delete(c.nulls, key)
	c.nullsMu.Unlock()
	if client := GetClient(); client != nil {
		_ = client.Del(key)
	}
}

var globalTokenCache *McpTokenCache
var tokenOnce sync.Once

func GetMcpTokenCache() *McpTokenCache {
	tokenOnce.Do(func() { globalTokenCache = NewMcpTokenCache() })
	return globalTokenCache
}
