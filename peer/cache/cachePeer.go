package cache

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/meguriri/GoCache/peer/callback"
	pb "github.com/meguriri/GoCache/peer/proto"
	"github.com/meguriri/GoCache/peer/replacement"
	"github.com/meguriri/GoCache/peer/replacement/manager"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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
	pb.UnimplementedGroupCacheServer
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

func (c *CachePeer) Listen() {
	log.Println("grpc server listen on:", c.Addr)
	lis, err := net.Listen("tcp", c.Addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return
	}
	// 实例化grpc服务端
	c.Server = grpc.NewServer()
	// 注册Greeter服务
	pb.RegisterGroupCacheServer(c.Server, c)
	// 往grpc服务端注册反射服务
	reflection.Register(c.Server)
	// 启动grpc服务
	if err := c.Server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	if _, ok := <-c.KillSignal; ok {
		fmt.Println("exit!")
		return
	}
}

func (c *CachePeer) Connect(ctx context.Context, in *pb.ConnectRequest) (*pb.ConnectResponse, error) {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	if in.Address != c.Addr {
		return &pb.ConnectResponse{Code: 404}, fmt.Errorf("[Invalid request]server address is not %s", in.Address)
	}
	log.Println("[connect] success!")
	c.Name, c.CacheBytes = in.Name, in.MaxBytes
	return &pb.ConnectResponse{Code: 200}, nil
}

func (c *CachePeer) Get(ctx context.Context, in *pb.GetRequest) (out *pb.GetResponse, err error) {

	key := in.GetKey()
	//验证key合法性
	if key == "" {
		return out, fmt.Errorf("key is required")
	}

	//从本地节点获取缓存值
	c.Lock.RLock()
	defer c.Lock.RUnlock()

	//cache未绑定底层缓存区
	if c.Manager == nil {
		return out, fmt.Errorf("[CachePeers Get] %s 未绑定底层缓存区", c.Addr)
	}

	//从底层缓存区获取缓存值
	if v, ok := c.Manager.Get(key); ok {
		log.Println("[Cache] hit")
		return &pb.GetResponse{Value: v.(ByteView).ToByte()}, nil
	} else if c.CallBackFunc != nil {
		view, err := c.storage(key)
		if err != nil {
			return out, err
		}
		return &pb.GetResponse{Value: view.ToByte()}, nil
	}
	return out, fmt.Errorf("[CachePeers Get] %s 缓存未命中", c.Addr)
}

func (c *CachePeer) Set(ctx context.Context, in *pb.SetRequest) (out *pb.SetResponse, err error) {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	key, value := in.Key, in.Value
	c.Manager.Add(key, ByteView(value))
	c.ModifyCnt++
	return &pb.SetResponse{Status: true}, nil
}

func (c *CachePeer) Del(ctx context.Context, in *pb.DelRequest) (out *pb.DelResponse, err error) {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	key := in.Key
	ok := c.Manager.Delete(key)
	if ok {
		c.ModifyCnt++
		return &pb.DelResponse{Status: ok}, nil
	}
	return &pb.DelResponse{Status: ok}, fmt.Errorf("del error")
}

func (c *CachePeer) Kill(ctx context.Context, in *pb.KillRequest) (*pb.KillResponse, error) {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	log.Println("kill reaseon: ", string(in.Reason))
	c.ticker.Stop()
	c.KillSignal <- struct{}{}
	return &pb.KillResponse{Status: true}, nil
}

func (c *CachePeer) GetAllCache(ctx context.Context, in *pb.GetAllCacheRequest) (*pb.GetAllCacheResponse, error) {
	c.Lock.RLock()
	defer c.Lock.RUnlock()
	res := c.Manager.GetAll()
	return &pb.GetAllCacheResponse{Size: int64(len(res)), Entry: res}, nil
}

func (c *CachePeer) Info(ctx context.Context, in *pb.InfoRequest) (*pb.InfoResponse, error) {
	c.Lock.RLock()
	defer c.Lock.RUnlock()
	return &pb.InfoResponse{
		Name:        c.Name,
		Address:     c.Addr,
		Replacement: replacement.ReplacementPolicy,
		UsedBytes:   int64(c.Manager.Len()),
		CacheBytes:  c.CacheBytes,
	}, nil
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
	bytes, _ := ioutil.ReadAll(buf)
	l := strings.Split(string(bytes), "\r\n")
	for _, v := range l {
		if v != "" {
			entry := strings.Split(v, ",")
			c.Manager.Add(entry[0], ByteView(entry[1]))
		}
	}
}
