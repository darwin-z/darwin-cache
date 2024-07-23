package core

import (
	"darwin-cache/core/lru"
	"sync"
)

// 并发控制,线程安全的访问
type cache struct {
	mu         sync.Mutex           // 互斥锁
	lru        *lru.LRUCacheManager // lru缓存
	cacheBytes int64                // 缓存最大容量
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
