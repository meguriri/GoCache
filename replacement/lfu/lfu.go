package lfu

import (
	"container/list"

	"github.com/meguriri/GoCache/data"
)

type lfuCache struct {
	maxBytes  int64                              //允许使用的最大内存
	nBytes    int64                              //当前使用的内存
	list      *list.List                         //双向链表
	cacheMap  map[string]*list.Element           //指向链表节点的字典
	OnEvicted func(key string, value data.Value) //节点被移除的回调函数
}

func New(onEvicted func(key string, value data.Value)) *lfuCache {
	return &lfuCache{}
}

func (c *lfuCache) Len() int {
	return 0
}

func (c *lfuCache) Get(string) (data.Value, bool) {
	return nil, false
}

func (c *lfuCache) RemoveOldest() {

}

func (c *lfuCache) Add(string, data.Value) {

}

func (c *lfuCache) GetAll() {

}
