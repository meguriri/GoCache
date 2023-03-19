package data

var (
	MaxBytes          int64
	ReplacementPolicy string
)

type Value interface { //键值对的值接口
	Len() int //value占的内存大小
}

type Entry struct { //双向链表中的节点数据类型
	Key   string //键
	Value Value  //值
}
