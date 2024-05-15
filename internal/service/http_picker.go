package service

import (
	"fmt"
	"leicache/internal/service/consistenthash"
	"leicache/utils/logger"
	"net/http"
	"strings"
	"sync"
)

// 在编译时检查类型安全，确保一个类型实现了某个接口
var _ Picker = (*HTTPPool)(nil)

const (
	defaultBasePath  = "/_leicache/"
	apiServerAddr    = "127.0.0.1:9999"
	defaultNumVisual = 50
)

type HTTPPool struct {
	address      string
	basePath     string
	mu           sync.Mutex
	peers        *consistenthash.Map
	httpFetchers map[string]*httpFetcher
}

func NewHTTPPool(address string) *HTTPPool {
	return &HTTPPool{
		address:  address,
		basePath: defaultBasePath,
	}
}

// Log HTTPPool implement HTTP Handler interface
func (p *HTTPPool) Log(format string, v ...interface{}) {
	logger.LogrusObj.Infof("[Server %s] %s", p.address, fmt.Sprintf(format, v...))
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}

	// print the requested method and requested resource path
	p.Log("%s %s", r.Method, r.URL.Path)

	// prefix/group/key
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request format, expected prefix/group/key", http.StatusBadRequest)
		return
	}
	groupName, key := parts[0], parts[1]

	g := GetGroup(groupName)
	if g == nil {
		http.Error(w, "no such group"+groupName, http.StatusBadRequest)
		return
	}

	view, err := g.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	// write value's deep copy
	_, err = w.Write(view.ByteSlice())
	if err != nil {
		return
	}
}

// Pick implementing Picker Interface.
// function: according to the specific key, select the node and return the HTTP client corresponding to the node.
func (p *HTTPPool) Pick(key string) (Fetcher, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	peerAddress := p.peers.GetTruthNode(key)
	if peerAddress == p.address {
		// upper layer get the value of the key locally after receiving false
		return nil, false
	}

	logger.LogrusObj.Infof("[dispatcher peer %s] pick remote peer: %s", apiServerAddr, peerAddress)
	return p.httpFetchers[peerAddress], true
}

// UpdatePeers rebuilding a consistent hash ring by new peers list
func (p *HTTPPool) UpdatePeers(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.peers = consistenthash.New(defaultNumVisual, nil)
	p.peers.Add(peers)
	p.httpFetchers = make(map[string]*httpFetcher, len(peers))

	for _, peer := range peers {
		p.httpFetchers[peer] = &httpFetcher{
			baseURL: peer + p.basePath, // such "http://10.0.0.1:9999/_leicache/"
		}
	}
}
