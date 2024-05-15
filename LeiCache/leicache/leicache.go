package leicache

import (
	"fmt"
	pb "leicache/leicachepb"
	"leicache/singleflight"
	"log"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name      string
	getter    Getter     // (3)
	mainCache cache      // (1)
	peers     PeerPicker // (2)
	// use singleflight.Group to make sure that
	// each key is only fetched once
	loader *singleflight.Group
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func NewGroup(name string, m int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: m},
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

// RegisterPeers 将 实现了 PeerPicker 接口的 HTTPPool 注入到 Group 中
// 对应group有http调用远程节点
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// new get cache store in g.mainCache
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	// (1) get from its own cache
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[LeiCache hit]")
		return v, nil
	}
	// cache not hit
	return g.load(key)
}

// 确保了并发场景下针对相同的 key，load 过程只会调用一次。
func (g *Group) load(key string) (value ByteView, err error) {
	// each key is only fetched once (either locally or remotely)
	// regardless of the number of concurrent callers.
	callOnce, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err := g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
			}
			log.Println("[LeiCache] Failed to get from peer", err)
		}

		return g.getLocally(key)
	})

	if err == nil {
		return callOnce.(ByteView), nil
	}
	return
}

func (g *Group) getLocally(key string) (ByteView, error) {
	// call user defined Getter
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

// 使用实现了 PeerGetter 接口的 httpGetter 访问远程节点，获取缓存值。
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}
