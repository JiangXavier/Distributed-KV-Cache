package service

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"leicache/internal/service/singleflight"
	"leicache/utils/logger"
	"sync"
	"time"
)

var (
	mu           sync.RWMutex
	GroupManager = make(map[string]*Group)
)

type Group struct {
	name      string
	cache     *cache                     // 缓存 (1)
	retriever Retriever                  // 回调数据库 (3)
	server    Picker                     // 获取远程节点 (2)
	flight    *singleflight.SingleFlight // 缓存击穿
}

func NewGroup(name string, strategy string, maxBytes int64, retriever Retriever) *Group {
	if retriever == nil {
		panic("Group Retriever must be existed!")
	}

	if _, ok := GroupManager[name]; ok {
		return GroupManager[name]
	}
	g := &Group{
		name:      name,
		cache:     newCache(strategy, maxBytes),
		retriever: retriever,
		flight:    singleflight.NewSingleFlight(10 * time.Second),
	}
	mu.Lock()
	GroupManager[name] = g
	mu.Unlock()
	return g
}

// RegisterServer register server Picker for group
func (g *Group) RegisterServer(p Picker) {
	if g.server != nil {
		panic("group had been registered server")
	}
	g.server = p
}

// GetGroup get the actual Group object
func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	return GroupManager[name]
}

//func DestroryGroup(name string) {
//	g := GetGroup(name)
//	if g != nil {
//		svr := g.server.(*Server)
//		svr.Stop()
//		delete(GroupManager, name)
//	}
//}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	// (1) get from the cache
	if value, ok := g.cache.get(key); ok {
		logger.LogrusObj.Infof("Group %s 缓存命中..., key %s...", g.name, key)
		return value, nil
	}

	// cache missing
	return g.load(key)
}

func (g *Group) load(key string) (ByteView, error) {
	// singleflight
	view, err := g.flight.Do(key, func() (interface{}, error) {
		if g.server != nil {
			if fetcher, ok := g.server.Pick(key); ok {
				bytes, err := fetcher.Fetch(g.name, key)
				if err == nil {
					return ByteView{b: cloneBytes(bytes)}, nil
				}
				logger.LogrusObj.Warnf("fetch key %s failed, error: %s\n", fetcher, err.Error())
			}
		}
		return g.getLocally(key)
	})

	if err == nil {
		return view.(ByteView), nil
	}

	return ByteView{}, err
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.retriever.retrieve(key)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.LogrusObj.Warnf("对于不存在的key，为了防止缓存穿透， 先存入缓存中并设置合理过期时间")
			g.cache.put(key, ByteView{})
		}
		return ByteView{}, err
	}

	value := ByteView{b: cloneBytes(bytes)}

	g.populateCache(key, value)

	return value, nil
}

// 从数据库获取数据后更新缓存
func (g *Group) populateCache(key string, value ByteView) {
	g.cache.put(key, value)
}
