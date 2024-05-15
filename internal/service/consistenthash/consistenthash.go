package consistenthash

import (
	"hash/crc32"
	"leicache/utils/logger"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

// Map hash cycle
type Map struct {
	hash      Hash           // hash function
	numVisual int            // number of the visual machines
	keys      []int          // all the hash keys
	mapping   map[int]string // map from the visual machine to actual machine
}

// New create a Map instance
func New(numVisual int, fn Hash) *Map {
	m := &Map{
		hash:      fn,
		numVisual: numVisual,
		mapping:   make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

func (m *Map) Add(keys []string) {
	for _, key := range keys {
		for i := 0; i < m.numVisual; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.mapping[hash] = key
		}
	}
	sort.Ints(m.keys)
}

func (m *Map) GetTruthNode(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	return m.mapping[m.keys[idx%len(m.keys)]]
}

func (m *Map) RemovePeer(peer string) {
	// 删除真实节点
	virtualHash := []int{}
	for key, v := range m.mapping {
		logger.LogrusObj.Warn("peers:", v)
		if v == peer {
			delete(m.mapping, key)
			virtualHash = append(virtualHash, key)
		}
	}

	for i := 0; i < len(virtualHash); i++ {
		for index, value := range m.keys {
			if virtualHash[i] == value {
				m.keys = append(m.keys[:index], m.keys[index+1:]...)
			}
		}
	}

}
