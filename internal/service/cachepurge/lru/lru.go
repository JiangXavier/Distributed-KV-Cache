package lru

import (
	"container/list"
	"leicache/config"
	"leicache/internal/service/interfaces"
	"leicache/utils/logger"
	"sync"
	"time"
)

type LRUCache struct {
	MaxBytes  int64
	usedByte  int64
	ll        *list.List
	cache     map[string]*list.Element
	mu        sync.RWMutex
	OnEvicted func(key string, value interfaces.Value)
}

func NewLRUCache(m int64, onevicted func(string, interfaces.Value)) *LRUCache {
	l := &LRUCache{
		MaxBytes:  m,
		usedByte:  0,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		mu:        sync.RWMutex{},
		OnEvicted: onevicted,
	}
	ttl := time.Duration(config.Conf.Services["leicache"].TTL) * time.Second

	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()

		for {
			<-ticker.C
			l.CleanUp(ttl)
			logger.LogrusObj.Warnf("触发过期缓存清理后台任务...")
		}
	}()

	return l
}

func (c *LRUCache) CleanUp(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for e := c.ll.Front(); e != nil; {
		next := e.Next()
		if entry, ok := e.Value.(*interfaces.Entry); ok && entry != nil {
			if entry.Expired(ttl) {
				c.ll.Remove(e)
				delete(c.cache, entry.Key)
				c.usedByte -= int64(len(entry.Key)) + int64(entry.Value.Len())
				// 失效了，回调
				if c.OnEvicted != nil {
					c.OnEvicted(entry.Key, entry.Value)
				}
			}
		}
		e = next
	}
}

func (c *LRUCache) Len() int {
	return c.ll.Len()
}

func (c *LRUCache) RemoveOldest() {
	c.mu.Lock()
	defer c.mu.Unlock()

	ele := c.ll.Front()
	if ele == nil {
		return
	}
	// 处理一下断言结果，而不是直接 panic
	kv, ok := c.ll.Remove(ele).(*interfaces.Entry)
	if !ok {
		logger.LogrusObj.Error("error: Item in LRU cache is not of type *interfaces.Entry")
		return
	}
	delete(c.cache, kv.Key)
	c.usedByte -= int64(len(kv.Key)) + int64(kv.Value.Len())

	if c.OnEvicted != nil {
		c.OnEvicted(kv.Key, kv.Value)
	}
}

func (c *LRUCache) Get(key string) (value interfaces.Value, updateAt *time.Time, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToBack(ele)
		e := ele.Value.(*interfaces.Entry)
		e.Touch()
		return e.Value, e.UpdateAt, ok
	}
	return
}

func (c *LRUCache) Put(key string, value interfaces.Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToBack(ele)
		kv := ele.Value.(*interfaces.Entry)
		kv.Touch()
		c.usedByte += int64(value.Len()) - int64(kv.Value.Len())
		kv.Value = value
	} else {
		kv := &interfaces.Entry{
			Key:      key,
			Value:    value,
			UpdateAt: nil,
		}
		kv.Touch()
		ele := c.ll.PushBack(kv)
		c.cache[key] = ele
		c.usedByte += int64(len(kv.Key)) + int64(kv.Value.Len())
	}
	for c.MaxBytes != 0 && c.MaxBytes < c.usedByte {
		c.RemoveOldest()
	}
}
