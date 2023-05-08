package cache

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/meguriri/GoCache/peer/callback"
	pb "github.com/meguriri/GoCache/peer/proto"
	"github.com/meguriri/GoCache/peer/replacement"
	"github.com/meguriri/GoCache/peer/replacement/manager"
	"google.golang.org/grpc"
)

var (
	SlaveAddress string
	SlavePort    string
	PeerAddress  string
	PeerPort     string
	SaveSeconds  int
	SaveModify   int
)

type CachePeer struct {
	pb.UnimplementedGoCacheServer
	Server       *grpc.Server
	Name         string
	Addr         string                   //cache的地址
	CallBackFunc callback.CallBack        //缓存未命中的回调函数
	Lock         sync.RWMutex             //互斥锁
	Manager      replacement.CacheManager //cache 底层存储和淘汰算法
	CacheBytes   int64                    //缓存字节最大值
	KillSignal   chan struct{}
	ticker       *time.Ticker
	ModifyCnt    int
}

func NewPeer(callback callback.CallBack) *CachePeer {

	//回调函数为空
	if callback == nil {
		log.Println("nil callBack func")
	}
	return &CachePeer{
		Addr:         PeerAddress + ":" + PeerPort,
		CallBackFunc: callback,
		Lock:         sync.RWMutex{},
		Manager:      manager.NewCache(replacement.ReplacementPolicy),
		CacheBytes:   replacement.MaxBytes,
		KillSignal:   make(chan struct{}),
	}
}

func (c *CachePeer) storage(key string) (ByteView, error) {
	log.Println("[Search call back] ")
	value, err := c.CallBackFunc.Get(key)
	if err != nil {
		return ByteView{}, fmt.Errorf("[Search call back] no local storage")
	}
	c.Manager.Add(key, ByteView(value))
	return ByteView(value), nil
}

func (c *CachePeer) Save() {
	c.ticker = time.NewTicker(time.Second * time.Duration(SaveSeconds))
	for range c.ticker.C {
		if c.ModifyCnt >= SaveModify {
			c.ModifyCnt = 0
			c.Snapshot()
		}
	}
	fmt.Println("save over")
}

func (c *CachePeer) Snapshot() {
	res := c.Manager.GetAll()
	file, err := os.Create("save.gdb")
	if err != nil {
		fmt.Println("文件创建失败 ", err.Error())
		return
	}
	defer file.Close()

	for _, v := range res {
		file.Write(v)
	}
}

func (c *CachePeer) ReadLocalStorage() {
	file, _ := os.Open("save.gdb")
	buf := bytes.NewBuffer([]byte{})
	buf.ReadFrom(file)
	bytes, _ := io.ReadAll(buf)
	l := strings.Split(string(bytes), "\r\n")
	for _, v := range l {
		if v != "" {
			entry := strings.Split(v, ",")
			c.Manager.Add(entry[0], ByteView(entry[1]))
		}
	}
}
