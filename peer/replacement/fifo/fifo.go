package fifo

import (
	"bytes"
	"container/list"

	"github.com/meguriri/GoCache/peer/replacement"
)

type fifoCacheManager struct { //Cache
	maxBytes  int64                                     //允许使用的最大内存
	nBytes    int64                                     //当前使用的内存
	list      *list.List                                //双向链表
	cacheMap  map[string]*list.Element                  //指向链表节点的字典
	OnEvicted func(key string, value replacement.Value) //节点被移除的回调函数
}

func New(onEvicted func(key string, value replacement.Value)) *fifoCacheManager { //初始化Cache
	return &fifoCacheManager{
		maxBytes:  replacement.MaxBytes,
		nBytes:    0,
		list:      list.New(),
		cacheMap:  make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *fifoCacheManager) Len() int {
	return int(c.nBytes)
}

func (c *fifoCacheManager) Get(key string) (replacement.Value, bool) {
	if element, ok := c.cacheMap[key]; ok {
		kv := element.Value.(*replacement.Entry)
		return kv.Value, true
	}
	return nil, false
}

func (c *fifoCacheManager) RemoveOldest() {
	if element := c.list.Front(); element != nil {
		c.list.Remove(element)
		kv := element.Value.(*replacement.Entry)
		delete(c.cacheMap, kv.Key)
		c.nBytes -= int64(len(kv.Key)) + int64(kv.Value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.Key, kv.Value)
		}
	}
}

func (c *fifoCacheManager) Add(key string, value replacement.Value) {
	if element, ok := c.cacheMap[key]; ok {
		kv := element.Value.(*replacement.Entry)
		c.nBytes = c.nBytes - int64(kv.Value.Len()) + int64(value.Len())
		kv.Value = value
	} else {
		element := c.list.PushBack(&replacement.Entry{Key: key, Value: value})
		c.cacheMap[key] = element
		c.nBytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}

func (c *fifoCacheManager) GetAll() [][]byte {
	res := make([][]byte, 0)
	for i := c.list.Front(); i != nil; i = i.Next() {
		kv := i.Value.(*replacement.Entry)
		buffer := bytes.NewBufferString(kv.Key + ",")
		buffer.Write(kv.Value.ToByte())
		buffer.WriteByte('\r')
		buffer.WriteByte('\n')
		bytes := buffer.Bytes()
		res = append(res, bytes)
	}
	return res
}

func (c *fifoCacheManager) Delete(key string) bool {
	if element, ok := c.cacheMap[key]; ok {
		kv := element.Value.(*replacement.Entry)
		c.list.Remove(element)
		delete(c.cacheMap, key)
		c.nBytes -= int64(len(kv.Key)) + int64(kv.Value.Len())
		return true
	}
	return false
}
