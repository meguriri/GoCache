package fifo

import (
	"container/list"
	"fmt"

	"github.com/meguriri/GoCache/data"
)

type fifoCacheManager struct { //Cache
	maxBytes  int64                              //允许使用的最大内存
	nBytes    int64                              //当前使用的内存
	list      *list.List                         //双向链表
	cacheMap  map[string]*list.Element           //指向链表节点的字典
	OnEvicted func(key string, value data.Value) //节点被移除的回调函数
}

func New(onEvicted func(key string, value data.Value)) *fifoCacheManager { //初始化Cache
	return &fifoCacheManager{
		maxBytes:  data.MaxBytes,
		nBytes:    0,
		list:      list.New(),
		cacheMap:  make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *fifoCacheManager) Len() int {
	return c.list.Len()
}

func (c *fifoCacheManager) Get(key string) (data.Value, bool) {
	if element, ok := c.cacheMap[key]; ok {
		kv := element.Value.(*data.Entry)
		return kv.Value, true
	}
	return nil, false
}

func (c *fifoCacheManager) RemoveOldest() {
	if element := c.list.Front(); element != nil {
		c.list.Remove(element)
		kv := element.Value.(*data.Entry)
		delete(c.cacheMap, kv.Key)
		c.nBytes -= int64(len(kv.Key)) + int64(kv.Value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.Key, kv.Value)
		}
	}
}

func (c *fifoCacheManager) Add(key string, value data.Value) {
	if element, ok := c.cacheMap[key]; ok {
		kv := element.Value.(*data.Entry)
		c.nBytes = c.nBytes - int64(kv.Value.Len()) + int64(value.Len())
		kv.Value = value
	} else {
		element := c.list.PushBack(&data.Entry{Key: key, Value: value})
		c.cacheMap[key] = element
		c.nBytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}

func (c *fifoCacheManager) GetAll() {
	fmt.Println("MaxBytes: ", c.maxBytes, ";nowUsedBytes: ", c.nBytes)
	fmt.Printf("[")
	for i := c.list.Front(); i != nil; i = i.Next() {
		kv := i.Value.(*data.Entry)
		fmt.Printf("key: %v,value: %v; ", kv.Key, kv.Value)
	}
	fmt.Printf("]\n")
}
