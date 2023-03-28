package cache

import (
	"sync"

	"github.com/meguriri/GoCache/replacement"
	"github.com/meguriri/GoCache/replacement/manager"
)

type Cache struct {
	lock       sync.Mutex
	manager    replacement.CacheManager
	cacheBytes int64
}

func (c *Cache) add(key string, value ByteView) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.manager == nil {
		c.manager = manager.NewCache(replacement.ReplacementPolicy)
	}
	c.manager.Add(key, value)
}

func (c *Cache) get(key string) (value ByteView, ok bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.manager == nil {
		return value, false
	}
	if v, ok := c.manager.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
