package manager

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/meguriri/GoCache/cache"
	"github.com/meguriri/GoCache/callback"
	"github.com/meguriri/GoCache/consistenthash"
	pb "github.com/meguriri/GoCache/proto"
	"github.com/meguriri/GoCache/replacement"
	"github.com/meguriri/GoCache/replacement/manager"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	DefaultbasePath string //默认节点url地址
	DefaultReplicas int    //虚拟节点个数
)

type Manager struct {
	lock            sync.RWMutex                //peers并发访问安全的读写锁
	cachePeers      map[string]*cache.CachePeer //peer名与peer指针的映射表
	localCache      replacement.CacheManager
	localcacheBytes int64
	hash            consistenthash.Map
}

func NewManager() *Manager {
	return &Manager{
		lock:            sync.RWMutex{},
		cachePeers:      make(map[string]*cache.CachePeer),
		localCache:      manager.NewCache(replacement.ReplacementPolicy),
		localcacheBytes: 1024 * 1024, //todo
		hash:            *consistenthash.New(DefaultReplicas, nil),
	}
}

// 生成一个新的peer
func (m *Manager) NewPeer(addr string, cacheBytes int64, callback callback.CallBack) *cache.CachePeer {

	//回调函数为空
	if callback == nil {
		log.Println("nil callBack func")
	}

	peer := &cache.CachePeer{
		Addr:         addr,
		CallBackFunc: callback,
		Lock:         sync.Mutex{},
		Manager:      manager.NewCache(replacement.ReplacementPolicy),
		CacheBytes:   cacheBytes,
	}

	//加写锁
	m.lock.Lock()
	defer m.lock.Unlock()

	m.hash.Add(peer.Addr)

	//http
	//go http.ListenAndServe(addr[7:], peer)

	//grpc
	go func() {
		lis, err := net.Listen("tcp", addr[7:])
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		// 实例化grpc服务端
		peer.Server = grpc.NewServer()
		// 注册Greeter服务
		pb.RegisterGroupCacheServer(peer.Server, peer)
		// 往grpc服务端注册反射服务
		reflection.Register(peer.Server)
		// 启动grpc服务
		if err := peer.Server.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	//生成的新group存入groups映射表
	m.cachePeers[addr] = peer

	return peer
}

// 根据peer名称获取peer
func (m *Manager) GetPeer(addr string) *cache.CachePeer {
	//加读锁
	m.lock.RLock()
	defer m.lock.RUnlock()

	//从groups中获取对应group
	peer := m.cachePeers[addr]

	return peer
}

func (m *Manager) Get(ctx context.Context, req *pb.CacheRequest) (*pb.CacheResponse, error) {
	key := req.GetKey()
	res := &pb.CacheResponse{}
	//use local cache
	if len(m.cachePeers) == 0 {
		if v, ok := m.localCache.Get(key); ok {
			res.Value = v.(cache.ByteView).GetByte()
			return res, nil
		}
		return res, fmt.Errorf("no cache")
	}

	//grpc
	addr := m.hash.Get(key)
	conn, err := grpc.Dial(addr[7:], grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return &pb.CacheResponse{}, err
	}
	// 延迟关闭连接
	defer conn.Close()

	// 初始化Greeter服务客户端
	c := pb.NewGroupCacheClient(conn)

	// 初始化上下文，设置请求超时时间为1秒
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// 延迟关闭请求会话
	defer cancel()

	// 调用Get接口，发送一条消息
	r, err := c.Get(ctx, &pb.CacheRequest{Key: key})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
		return &pb.CacheResponse{}, err
	}

	// 打印服务的返回的消息
	log.Printf("Greeting: %s", r.Value)
	return r, nil
}
