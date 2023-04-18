package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// 哈希函数类型
type Hash func(data []byte) uint32

type Map struct {
	hash     Hash           //Hash函数
	replicas int            //虚拟节点倍数
	keys     []int          //哈希环
	hashMap  map[int]string //虚拟节点与真实节点映射表
}

// 创建一致性哈希
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// 添加节点
func (m *Map) Add(peers ...string) {
	//将真实节点和虚拟节点添加到一致性哈希环上
	for _, peer := range peers {
		for i := 0; i < m.replicas; i++ {

			//获取哈希值
			hash := int(m.hash([]byte(strconv.Itoa(i) + peer)))

			//将哈希值添加到一致性哈希环上
			m.keys = append(m.keys, hash)

			//添加虚拟或真实节点哈希值到真实节点名的映射关系
			m.hashMap[hash] = peer
		}
	}

	//哈希值排序，让一致性哈希环有序
	sort.Ints(m.keys)
}

// 从一致性哈希环上获取节点值
func (m *Map) Get(key string) string {

	//验证key合法性
	if len(m.keys) == 0 {
		return ""
	}

	//获取key的哈希值
	hash := int(m.hash([]byte(key)))

	//寻找第一个大于等于的节点的序号
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	//得到哈希值后从映射中返回节点值，当idx==len(m.keys)时，取余处理从m.keys[0]中获得哈希值
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
