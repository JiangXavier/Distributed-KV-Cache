package lru

import "container/list"

type Cache struct {
	MaxBytes  int64
	usedByte  int64
	ll        *list.List
	cache     map[string]*list.Element
	OnEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func (c *Cache) Len() int {
	return c.ll.Len()
}

func New(m int64, evicted func(key string, value Value)) *Cache {
	return &Cache{
		MaxBytes:  m,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: evicted,
	}
}

func (c *Cache) Add(key string, value Value) {
	if node, ok := c.cache[key]; ok {
		c.ll.MoveToFront(node)
		c.usedByte += int64(value.Len()) - int64(node.Value.(*entry).value.Len())
		node.Value.(*entry).value = value
	} else {
		node = c.ll.PushFront(&entry{key, value})
		c.cache[key] = node
		c.usedByte += int64(len(key)) + int64(value.Len())
	}
	for c.MaxBytes != 0 && c.usedByte > c.MaxBytes {
		c.RemoveOld()
	}
}

func (c *Cache) RemoveOld() {
	node := c.ll.Back()
	if node != nil {
		c.ll.Remove(node)
		delete(c.cache, node.Value.(*entry).key)
		c.usedByte -= int64(len(node.Value.(*entry).key)) + int64(node.Value.(*entry).value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(node.Value.(*entry).key, node.Value.(*entry).value)
		}
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if node, ok := c.cache[key]; ok {
		c.ll.MoveToFront(node)
		return node.Value.(*entry).value, ok
	}
	return
}
