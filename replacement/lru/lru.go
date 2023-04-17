package lru

import (
	"container/list"
	"fmt"

	"github.com/meguriri/GoCache/replacement"
)

type lruCacheManager struct { //Cache
	maxBytes  int64                                     //允许使用的最大内存
	nBytes    int64                                     //当前使用的内存
	list      *list.List                                //双向链表
	cacheMap  map[string]*list.Element                  //指向链表节点的字典
	OnEvicted func(key string, value replacement.Value) //节点被移除的回调函数
}

func New(onEvicted func(key string, value replacement.Value)) *lruCacheManager { //初始化Cache
	return &lruCacheManager{
		maxBytes:  replacement.MaxBytes,
		nBytes:    0,
		list:      list.New(),
		cacheMap:  make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *lruCacheManager) Len() int { //获取链表长度
	return c.list.Len()
}

func (c *lruCacheManager) Get(key string) (replacement.Value, bool) { //获得value
	if element, ok := c.cacheMap[key]; ok { //key存在
		c.list.MoveToBack(element)               //将节点移至队尾
		kv := element.Value.(*replacement.Entry) //获取该键值对
		return kv.Value, true                    //返回value，true
	}
	return nil, false //key不存在，返回nil，false
}

func (c *lruCacheManager) RemoveOldest() { //缓存淘汰，删除队首节点
	if element := c.list.Front(); element != nil { //队首存在元素
		c.list.Remove(element)                                 //链表中删除队首节点
		kv := element.Value.(*replacement.Entry)               //获取队首键值对
		delete(c.cacheMap, kv.Key)                             //在字典中删除该节点映射关系
		c.nBytes -= int64(len(kv.Key)) + int64(kv.Value.Len()) //更新使用内存大小
		if c.OnEvicted != nil {                                //回调函数不为空，调用回调函数
			c.OnEvicted(kv.Key, kv.Value)
		}
	}
}

func (c *lruCacheManager) Add(key string, value replacement.Value) { //添加或更新节点到cache中
	if element, ok := c.cacheMap[key]; ok { //节点存在，更新
		c.list.MoveToBack(element)                                       //节点移植队尾
		kv := element.Value.(*replacement.Entry)                         //获取键值对
		c.nBytes = c.nBytes - int64(kv.Value.Len()) + int64(value.Len()) //更新使用内存大小
		kv.Value = value                                                 //更新value
	} else { //节点不存在，添加
		element := c.list.PushBack(&replacement.Entry{Key: key, Value: value}) //新节点添加到队尾
		c.cacheMap[key] = element                                              //新节点添加到字典中
		c.nBytes += int64(len(key)) + int64(value.Len())                       //更新使用内存大小
	}
	for c.maxBytes != 0 && c.maxBytes < c.nBytes { //添加后超过最大内存
		c.RemoveOldest() //删除队首节点
	}
}

func (c *lruCacheManager) GetAll() { //获取全部节点
	fmt.Println("MaxBytes: ", c.maxBytes, ";nowUsedBytes: ", c.nBytes)
	fmt.Printf("[")
	for i := c.list.Front(); i != nil; i = i.Next() {
		kv := i.Value.(*replacement.Entry)
		fmt.Printf("key: %v,value: %v; ", kv.Key, kv.Value)
	}
	fmt.Printf("]\n\n")
}

func (c *lruCacheManager) Delete(key string) bool {
	if element, ok := c.cacheMap[key]; ok {
		kv := element.Value.(*replacement.Entry)
		c.list.Remove(element)
		delete(c.cacheMap, key)
		c.nBytes -= int64(len(kv.Key)) + int64(kv.Value.Len())
		return true
	}
	return false
}
