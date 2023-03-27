package data

var (
	MaxBytes          int64
	ReplacementPolicy string
)

type CacheManager interface { //Cache
	Len() int
	Get(string) (Value, bool)
	RemoveOldest()
	Add(string, Value)
	GetAll()
}

type Value interface { //键值对的值接口
	Len() int //value占的内存大小
}

type Entry struct { //双向链表中的节点数据类型
	Key   string //键
	Value Value  //值
}

type LfuEntry struct { //LFU使用的节点数据类型
	Key   string //键
	Value Value  //值
	Used  int    //使用频率
}
