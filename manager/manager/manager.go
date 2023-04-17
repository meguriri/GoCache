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
	addr            string
	lock            sync.RWMutex                //peers并发访问安全的读写锁
	cachePeers      map[string]*cache.CachePeer //peer名与peer指针的映射表
	localCache      replacement.CacheManager
	localcacheBytes int64
	callback        callback.CallBack
	hash            consistenthash.Map
}

func (m Manager) GetAddr() string {
	return m.addr
}

func NewManager(addr string, cacheBytes int64, callback callback.CallBack) *Manager {
	return &Manager{
		addr:            addr,
		lock:            sync.RWMutex{},
		cachePeers:      make(map[string]*cache.CachePeer),
		localCache:      manager.NewCache(replacement.ReplacementPolicy),
		localcacheBytes: cacheBytes,
		callback:        callback,
		hash:            *consistenthash.New(DefaultReplicas, nil),
	}
}

// 生成一个新的peer
func (m *Manager) NewPeer(name, addr string, cacheBytes int64, callback callback.CallBack) *cache.CachePeer {

	//回调函数为空
	if callback == nil {
		log.Println("nil callBack func")
	}

	peer := &cache.CachePeer{
		Name:         name,
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

	//grpc
	go func() {
		lis, err := net.Listen("tcp", addr)
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

func (m *Manager) GetAllPeer() []*cache.CachePeer {
	l := make([]*cache.CachePeer, 0)
	for _, c := range m.cachePeers {
		l = append(l, c)
	}
	return l
}

func (m *Manager) KillPeer(addr string) (bool, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return false, err
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
	r, err := c.Ping(ctx, &pb.PingRequest{Msg: "kill"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
		return false, err
	}

	// 打印服务的返回的消息
	log.Printf("Greeting: %s", r.Code)
	delete(m.cachePeers, addr)
	return true, nil
}

func (m *Manager) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	key := req.GetKey()
	res := &pb.GetResponse{}
	//use local cache
	if len(m.cachePeers) == 0 {
		if v, ok := m.localCache.Get(key); ok {
			log.Println("[Cache] hit")
			res.Value = v.(cache.ByteView).GetByte()
			return res, nil
		} else if m.callback != nil {
			view, err := m.storage(key)
			if err != nil {
				return &pb.GetResponse{}, err
			}
			return &pb.GetResponse{Value: view.GetByte()}, nil
		}
		return res, fmt.Errorf("no cache")
	}

	//grpc
	addr := m.hash.Get(key)
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return &pb.GetResponse{}, err
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
	r, err := c.Get(ctx, &pb.GetRequest{Key: key})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
		return &pb.GetResponse{}, err
	}

	// 打印服务的返回的消息
	log.Printf("Greeting: %s", r.Value)
	return r, nil
}

func (m *Manager) storage(key string) (cache.ByteView, error) {
	log.Println("[Search call back] ")
	value, err := m.callback.Get(key)
	if err != nil {
		return cache.ByteView{}, fmt.Errorf("[Search call back] no local storage")
	}
	m.localCache.Add(key, cache.ByteView(value))
	return cache.ByteView(value), nil
}

func (m *Manager) Set(req *pb.SetRequest) (*pb.SetResponse, error) {
	key, value := req.GetKey(), req.GetValue()
	if len(m.cachePeers) == 0 {
		res := &pb.SetResponse{}
		m.localCache.Add(key, cache.ByteView(value))
		res.Status = true
		return res, nil
	}
	addr := m.hash.Get(key)
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return &pb.SetResponse{Status: false}, err
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
	r, err := c.Set(ctx, &pb.SetRequest{Key: key, Value: value})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
		return &pb.SetResponse{Status: false}, err
	}
	// 打印服务的返回的消息
	log.Printf("Greeting: %v", r.Status)
	return r, nil
}

func (m *Manager) Del(req *pb.DelRequest) (*pb.DelResponse, error) {
	key := req.GetKey()
	if len(m.cachePeers) == 0 {
		res := &pb.DelResponse{}
		m.localCache.Delete(key)
		res.Status = true
		return res, nil
	}
	addr := m.hash.Get(key)
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return &pb.DelResponse{Status: false}, err
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
	r, err := c.Del(ctx, &pb.DelRequest{Key: key})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
		return &pb.DelResponse{Status: false}, err
	}
	// 打印服务的返回的消息
	log.Printf("Greeting: %v", r.Status)
	return r, nil
}
