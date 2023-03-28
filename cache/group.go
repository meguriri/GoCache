package cache

import (
	"fmt"
	"log"
	"sync"

	"github.com/meguriri/GoCache/callback"
)

type Group struct {
	name         string
	callBackFunc callback.CallBack
	mainCache    Cache
}

var (
	lock   sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, callback callback.CallBack) *Group {
	if callback == nil {
		fmt.Println("nil callBack func")
	}
	lock.Lock()
	defer lock.Unlock()
	g := &Group{
		name:         name,
		callBackFunc: callback,
		mainCache:    Cache{lock: sync.Mutex{}, cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	lock.RLock()
	defer lock.RUnlock()
	g := groups[name]
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[Cache] hit")
		return v, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.callBackFunc.Get(key)
	log.Println("[getLocally] values ", string(bytes))
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value) //存入缓存中
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
	log.Println("[populateCache] add ok")
}
