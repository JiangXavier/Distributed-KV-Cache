package service

import (
	"sync"

	"leicache/utils/logger"

	"leicache/internal/service/cachepurge"
	"leicache/internal/service/cachepurge/interfaces"
)

// cache 模块负责提供对lru模块的并发控制

// 给 lru 上层并发上一层锁
type cache struct {
	mu           sync.RWMutex
	strategy     interfaces.CacheStrategy
	maxCacheSize int64 // 保证 lru 一定初始化
}

func newCache(strategy string, cacheSize int64) *cache {
	onEvicted := func(key string, val interfaces.Value) {
		logger.LogrusObj.Infof("缓存条目 [%s:%s] 被淘汰", key, val)
	}

	return &cache{
		maxCacheSize: cacheSize,
		strategy:     cachepurge.New(strategy, cacheSize, onEvicted),
	}
}

// 并发控制
func (c *cache) set(key string, value ByteView) {
	c.mu.Lock()
	c.strategy.Put(key, value)
	c.mu.Unlock()
}

func (c *cache) get(key string) (ByteView, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if v, _, ok := c.strategy.Get(key); ok { // Get 返回值是 Value 接口，直接类型断言
		return v.(ByteView), true
	} else {
		return ByteView{}, false
	}
}

func (c *cache) put(key string, val ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()

	logger.LogrusObj.Infof("存入数据库之后压入缓存, (key, value)=(%s, %s)", key, val)
	c.strategy.Put(key, val)
}
