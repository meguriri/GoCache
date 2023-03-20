package lfu

import (
	"container/list"
	"fmt"

	"github.com/meguriri/GoCache/data"
)

type lfuCache struct {
	maxBytes  int64                              //允许使用的最大内存
	nBytes    int64                              //当前使用的内存
	minFreq   int                                //最少使用频率
	cacheMap  map[string]*list.Element           //指向链表节点的字典
	freqMap   map[int]*list.List                 //使用频率的字典
	OnEvicted func(key string, value data.Value) //节点被移除的回调函数
}

func New(onEvicted func(key string, value data.Value)) *lfuCache {
	return &lfuCache{
		maxBytes:  data.MaxBytes,
		nBytes:    0,
		minFreq:   1,
		cacheMap:  make(map[string]*list.Element),
		freqMap:   make(map[int]*list.List),
		OnEvicted: onEvicted,
	}
}

func (c *lfuCache) Len() int {
	return len(c.cacheMap)
}

func (c *lfuCache) AddToNewFreqList(lfuEntry *data.LfuEntry, used int) *list.Element {
	var newList *list.List
	if l, ok := c.freqMap[used]; ok {
		newList = l

	} else {
		newList = new(list.List)
		c.freqMap[used] = newList
	}
	return newList.PushBack(lfuEntry)
}
func (c *lfuCache) RemoveFreqList(element *list.Element) *data.LfuEntry {
	kv := element.Value.(*data.LfuEntry)
	oldList := c.freqMap[kv.Used]
	oldList.Remove(element)
	if oldList.Len() == 0 {
		if kv.Used == c.minFreq {
			c.minFreq++
		}
		delete(c.freqMap, kv.Used)
	}
	kv.Used++
	// lfuEntry := &data.LfuEntry{
	// 	Key:   kv.Key,
	// 	Value: kv.Value,
	// 	Used:  kv.Used,
	// }
	//c.AddToNewFreqList(lfuEntry, kv.Used)
	return kv
}

func (c *lfuCache) Get(key string) (data.Value, bool) {
	if element, ok := c.cacheMap[key]; ok { //根据cacheMap 找到*element
		kv := c.RemoveFreqList(element)
		newElement := c.AddToNewFreqList(kv, kv.Used)
		c.cacheMap[key] = newElement
		return kv.Value, true
	}
	return nil, false
}

func (c *lfuCache) RemoveOldest() {
	oldList := c.freqMap[c.minFreq]
	element := oldList.Front()
	kv := element.Value.(*data.LfuEntry)
	delete(c.cacheMap, kv.Key)
	oldList.Remove(element)
	if oldList.Len() == 0 {
		delete(c.freqMap, c.minFreq)
		c.minFreq++
	}
	c.nBytes -= int64(len(kv.Key)) + int64(kv.Value.Len()) + 8
	if c.OnEvicted != nil { //回调函数不为空，调用回调函数
		c.OnEvicted(kv.Key, kv.Value)
	}
}

func (c *lfuCache) Add(key string, value data.Value) {
	if element, ok := c.cacheMap[key]; ok { //节点存在，更新  从cacheMap获取*element
		kv := c.RemoveFreqList(element)
		newElement := c.AddToNewFreqList(kv, kv.Used)
		c.cacheMap[key] = newElement
		c.nBytes = c.nBytes - int64(kv.Value.Len()) + int64(value.Len())
		kv.Value = value
	} else { //节点不存在，添加
		lfuEntry := &data.LfuEntry{
			Key:   key,
			Value: value,
			Used:  1,
		}
		c.minFreq = 1
		element := c.AddToNewFreqList(lfuEntry, 1)
		c.cacheMap[key] = element
		c.nBytes += int64(len(key)) + int64(value.Len()) + 8
		for c.maxBytes != 0 && c.maxBytes < c.nBytes {
			c.RemoveOldest()
		}

	}
}

func (c *lfuCache) GetAll() {
	fmt.Println("MaxBytes: ", c.maxBytes, ";nowUsedBytes: ", c.nBytes, ";minFreq: ", c.minFreq)
	fmt.Printf("{\n")
	for i, list := range c.freqMap {
		fmt.Printf("%d: [", i)
		for j := list.Front(); j != nil; j = j.Next() {
			kv := j.Value.(*data.LfuEntry)
			fmt.Printf("key:%v,value:%v; ", kv.Key, kv.Value)
		}
		fmt.Printf("]\n")
	}
	fmt.Printf("}\n\n")
}
