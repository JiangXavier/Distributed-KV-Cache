package cachepurge

import (
	"strings"

	"leicache/internal/service/cachepurge/fifo"

	"leicache/internal/service/cachepurge/interfaces"

	"leicache/internal/service/cachepurge/lfu"

	"leicache/internal/service/cachepurge/lru"
)

func New(name string, maxBytes int64, onEvicted func(string, interfaces.Value)) interfaces.CacheStrategy {
	name = strings.ToLower(name)
	switch name {
	case "fifo":
		return fifo.NewFIFOCache(maxBytes, onEvicted)
	case "lru":
		return lru.NewLRUCache(maxBytes, onEvicted)
	case "lfu":
		return lfu.NewLFUCache(maxBytes, onEvicted)
	}
	return nil
}
