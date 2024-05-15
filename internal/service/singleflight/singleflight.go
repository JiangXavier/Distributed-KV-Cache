package singleflight

import (
	"leicache/utils/logger"
	"sync"
	"time"
)

type Call struct {
	wg    sync.WaitGroup
	value interface{}
	err   error
}

type SingleFlight struct {
	mu     sync.RWMutex
	cache  map[string]*cachedValue
	m      map[string]*Call
	ttl    time.Duration
	ticker *time.Ticker
}

type cachedValue struct {
	value   interface{}
	expires time.Time
}

func NewSingleFlight(ttl time.Duration) *SingleFlight {
	sf := &SingleFlight{
		cache: make(map[string]*cachedValue),
		m:     make(map[string]*Call),
		ttl:   0,
	}
	sf.ticker = time.NewTicker(ttl)
	go sf.cacheCleaner()
	return sf
}

func (sf *SingleFlight) cacheCleaner() {
	for range sf.ticker.C {
		sf.mu.Lock()
		for key, cv := range sf.cache {
			if time.Now().After(cv.expires) {
				delete(sf.cache, key)
			}
		}
		sf.mu.Unlock()
	}
}

func (sf *SingleFlight) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	sf.mu.RLock()
	if cv, ok := sf.cache[key]; ok && time.Now().Before(cv.expires) {
		sf.mu.RUnlock()
		return cv.value, nil
	}

	c, ok := sf.m[key]
	sf.mu.RUnlock()

	if ok {
		logger.LogrusObj.Warnf("%s 已经在查询了，阻塞等待 goroutine 返回结果", key)
		c.wg.Wait()
		// 用于查询的 goroutine 已经返回，结果值已经存入 Call 结构体中
		return c.value, c.err
	}
	c = new(Call)
	c.wg.Add(1)

	sf.mu.Lock()
	sf.m[key] = c
	sf.mu.Unlock()
	// 开启查询，c.value 和 c.err 接收返回值
	c.value, c.err = fn()
	c.wg.Done()

	sf.mu.Lock()
	delete(sf.m, key)
	// 缓存结果
	if c.err == nil {
		sf.cache[key] = &cachedValue{
			value:   c.value,
			expires: time.Now().Add(sf.ttl),
		}
	}
	sf.mu.Unlock()

	return c.value, c.err
}
