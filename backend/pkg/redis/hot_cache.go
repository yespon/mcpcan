package redis

import (
	"sync"
	"time"
)

// Hot cache module encapsulates local hotspot caching logic and is parameter-controlled.

const (
	// LocalHotCacheTTL is TTL for local hot cache items
	LocalHotCacheTTL = 10 * time.Second
	// MaxLocalCacheSize is the max number of local hot cache items
	MaxLocalCacheSize = 1000
	// HotCleanInterval is the cleaning interval for hot cache
	HotCleanInterval = 30 * time.Second
)

var (
	hotCacheEnabled bool
	hotCacheMu      sync.RWMutex
	hotCache        = make(map[string]*localCacheItem)
	hotCacheStopCh  chan struct{}
	hotCacheWG      sync.WaitGroup
)

// localCacheItem represents a local hot cache item
type localCacheItem struct {
	info      *CacheInstanceInfo
	expireAt  time.Time
	accessCnt int64
}

// EnableHotCache enables the local hot cache.
// It is idempotent: if already enabled, it returns immediately.
// When enabling, it starts the cleaner goroutine only once.
func EnableHotCache() {
	if hotCacheEnabled {
		return
	}
	hotCacheEnabled = true
	if hotCacheStopCh == nil {
		hotCacheStopCh = make(chan struct{})
		hotCacheWG.Add(1)
		go hotCacheCleaner()
	}
}

// DisableHotCache disables the local hot cache.
// It is idempotent: if already disabled, it returns immediately.
// When disabling, it stops the cleaner goroutine and clears the cache.
func DisableHotCache() {
	if !hotCacheEnabled {
		return
	}
	hotCacheEnabled = false
	if hotCacheStopCh != nil {
		close(hotCacheStopCh)
		hotCacheWG.Wait()
		hotCacheStopCh = nil
	}
	ClearAllHotCache()
}

// GetHotCache returns cached info when enabled, otherwise nil.
func GetHotCache(key string) *CacheInstanceInfo {
	if !hotCacheEnabled {
		return nil
	}
	hotCacheMu.RLock()
	item, exists := hotCache[key]
	hotCacheMu.RUnlock()
	if !exists {
		return nil
	}
	if time.Now().Before(item.expireAt) {
		item.accessCnt++
		return item.info
	}
	// expired: let cleaner remove; return nil
	return nil
}

// SetHotCache sets a hotspot cache item when enabled.
func SetHotCache(key string, info *CacheInstanceInfo) {
	if !hotCacheEnabled || info == nil {
		return
	}
	hotCacheMu.Lock()
	// Simple capacity guard: random eviction will happen in cleaner if oversized.
	hotCache[key] = &localCacheItem{
		info:      info,
		expireAt:  time.Now().Add(LocalHotCacheTTL),
		accessCnt: 1,
	}
	hotCacheMu.Unlock()
}

// ClearHotCache removes one item from the hot cache.
func ClearHotCache(key string) {
	hotCacheMu.Lock()
	delete(hotCache, key)
	hotCacheMu.Unlock()
}

// ClearAllHotCache clears all items from the hot cache.
func ClearAllHotCache() {
	hotCacheMu.Lock()
	hotCache = make(map[string]*localCacheItem)
	hotCacheMu.Unlock()
}

// HotCacheSize returns number of items in the hot cache.
func HotCacheSize() int {
	hotCacheMu.RLock()
	size := len(hotCache)
	hotCacheMu.RUnlock()
	return size
}

// hotCacheCleaner periodically evicts expired and least-used items.
func hotCacheCleaner() {
	defer hotCacheWG.Done()
	ticker := time.NewTicker(HotCleanInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// Expired eviction
			now := time.Now()
			hotCacheMu.Lock()
			for key, item := range hotCache {
				if now.After(item.expireAt) {
					delete(hotCache, key)
				}
			}
			// Size-based eviction: remove least-used items if oversized
			if len(hotCache) > MaxLocalCacheSize {
				evictLeastUsed()
			}
			hotCacheMu.Unlock()
		case <-hotCacheStopCh:
			return
		}
	}
}

// evictLeastUsed removes least-used items up to half the cache size.
func evictLeastUsed() {
	type cacheItem struct {
		key  string
		item *localCacheItem
	}
	items := make([]cacheItem, 0, len(hotCache))
	for key, item := range hotCache {
		items = append(items, cacheItem{key: key, item: item})
	}
	// selection-like partial sort for bottom half
	for i := 0; i < len(items)/2; i++ {
		minIdx := i
		for j := i + 1; j < len(items); j++ {
			if items[j].item.accessCnt < items[minIdx].item.accessCnt {
				minIdx = j
			}
		}
		if minIdx != i {
			items[i], items[minIdx] = items[minIdx], items[i]
		}
	}
	cleanCount := len(items) / 2
	if cleanCount > MaxLocalCacheSize/2 {
		cleanCount = MaxLocalCacheSize / 2
	}
	for i := 0; i < cleanCount; i++ {
		delete(hotCache, items[i].key)
	}
}
