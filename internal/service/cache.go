package service

import (
	"leicache/internal/service/cachepurge"
	"leicache/internal/service/interfaces"
	"leicache/utils/logger"
	"sync"
)

// cache 模块负责提供对缓存淘汰模块的并发控制

type cache struct {
	mu           sync.RWMutex
	strategy     interfaces.CacheStrategy
	maxCacheSize int64
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
