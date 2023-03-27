package cache

import (
	"sync"

	"github.com/meguriri/GoCache/data"
	"github.com/meguriri/GoCache/replacement"
)

type cache struct {
	lock       sync.Mutex
	manager    data.CacheManager
	cacheBytes int64
}

func (c *cache) add(key string, value data.ByteView) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.manager == nil {
		c.manager = replacement.NewCache(data.ReplacementPolicy)
	}
	c.manager.Add(key, value)
}

func (c *cache) get(key string) (value data.ByteView, ok bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.manager == nil {
		return value, false
	}
	if v, ok := c.manager.Get(key); ok {
		return v.(data.ByteView), ok
	}
	return
}
