package manager

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/meguriri/GoCache/server/cache"
	"github.com/meguriri/GoCache/server/callback"
	"github.com/meguriri/GoCache/server/consistenthash"
	pb "github.com/meguriri/GoCache/server/proto"
	"github.com/meguriri/GoCache/server/replacement"
	"github.com/meguriri/GoCache/server/replacement/manager"
	"google.golang.org/grpc"
)

var (
	ManagerIP       string
	ManagerPort     string
	DefaultReplicas int //虚拟节点个数
)

type Manager struct {
	addr            string
	lock            sync.RWMutex                //peers并发访问安全的读写锁
	cachePeers      map[string]*grpc.ClientConn //peer名与grpc conn的映射表
	localCache      replacement.CacheManager
	localcacheBytes int64
	callback        callback.CallBack
	hash            consistenthash.Map
}

func NewManager(callback callback.CallBack) *Manager {
	return &Manager{
		addr:            ManagerIP + ":" + ManagerPort,
		lock:            sync.RWMutex{},
		cachePeers:      make(map[string]*grpc.ClientConn),
		localCache:      manager.NewCache(replacement.ReplacementPolicy),
		localcacheBytes: replacement.MaxBytes,
		callback:        callback,
		hash:            *consistenthash.New(DefaultReplicas, nil),
	}
}

// 生成一个新的peer
func (m *Manager) Connect(name, addr string, cacheBytes int64) bool {

	//加写锁
	m.lock.Lock()
	defer m.lock.Unlock()

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Printf("did not connect: %v\n", err)
		return false
	}

	// 初始化Greeter服务客户端
	client := pb.NewGroupCacheClient(conn)

	// 初始化上下文，设置请求超时时间为1秒
	ctx := context.Background()
	// 调用Get接口，发送一条消息
	r, err := client.Connect(ctx, &pb.ConnectRequest{Name: name, Address: addr, MaxBytes: cacheBytes})
	if err != nil {
		log.Printf("could not greet: %v\n", err)
		return false
	}

	// 打印服务的返回的消息
	log.Printf("Greeting: %d", r.Code)

	//生成的新group存入groups映射表
	if r.Code == 200 {
		m.hash.Add(addr)
		m.cachePeers[addr] = conn
	} else {
		log.Printf("grpc connect err code: %d", r.Code)
		return false
	}
	return true
}

// 根据peer名称获取peer
func (m *Manager) GetPeer(addr string) *grpc.ClientConn {
	//加读锁
	m.lock.RLock()
	defer m.lock.RUnlock()

	//从groups中获取对应group
	conn, ok := m.cachePeers[addr]
	if !ok {
		return nil
	}

	return conn
}

func (m *Manager) GetAllPeer() []*grpc.ClientConn {
	m.lock.RLock()
	defer m.lock.RUnlock()
	l := make([]*grpc.ClientConn, 0)
	for _, c := range m.cachePeers {
		l = append(l, c)
	}
	return l
}

func (m *Manager) GetAllPeerAddress() []string {
	m.lock.RLock()
	defer m.lock.RUnlock()
	l := make([]string, 0)
	for addr := range m.cachePeers {
		l = append(l, addr)
	}
	return l
}

func (m *Manager) Get(ctx context.Context, key string) ([]byte, error) {
	//use local cache
	if len(m.cachePeers) == 0 {
		if v, ok := m.localCache.Get(key); ok {
			log.Println("[Cache] hit")
			return v.(cache.ByteView).ToByte(), nil
		} else if m.callback != nil {
			view, err := m.storage(key)
			if err != nil {
				return []byte{}, err
			}
			return view.ToByte(), nil
		}
		return []byte{}, fmt.Errorf("no cache")
	}
	//grpc
	m.lock.RLock()
	defer m.lock.RUnlock()
	addr := m.hash.Get(key)
	conn, ok := m.cachePeers[addr]
	if !ok {
		return []byte{}, fmt.Errorf("%s,that does not exist", addr)
	}
	client := pb.NewGroupCacheClient(conn)
	// 调用Get接口，发送一条消息
	r, err := client.Get(ctx, &pb.GetRequest{Key: key})
	if err != nil {
		log.Printf("could not greet: %v\n", err)
		return []byte{}, err
	}

	// 打印服务的返回的消息
	log.Printf("Greeting: %s", r.Value)
	return r.GetValue(), nil
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

func (m *Manager) Set(ctx context.Context, key string, value []byte) bool {
	if len(m.cachePeers) == 0 {
		m.localCache.Add(key, cache.ByteView(value))
		return true
	}

	m.lock.RLock()
	defer m.lock.RUnlock()
	addr := m.hash.Get(key)
	conn, ok := m.cachePeers[addr]
	if !ok {
		return false
	}
	client := pb.NewGroupCacheClient(conn)

	// 调用Get接口，发送一条消息
	r, err := client.Set(ctx, &pb.SetRequest{Key: key, Value: value})
	if err != nil {
		log.Printf("could not greet: %v\n", err)
		return false
	}
	// 打印服务的返回的消息
	log.Printf("Greeting: %v", r.Status)
	return r.Status
}

func (m *Manager) Del(ctx context.Context, key string) bool {
	if len(m.cachePeers) == 0 {
		ok := m.localCache.Delete(key)
		return ok
	}

	m.lock.RLock()
	defer m.lock.RUnlock()
	addr := m.hash.Get(key)
	conn, ok := m.cachePeers[addr]
	if !ok {
		return false
	}
	client := pb.NewGroupCacheClient(conn)

	// 调用Get接口，发送一条消息
	r, err := client.Del(ctx, &pb.DelRequest{Key: key})
	if err != nil {
		log.Printf("could not greet: %v\n", err)
		return false
	}
	// 打印服务的返回的消息
	log.Printf("Greeting: %v", r.Status)
	return r.Status
}

func (m *Manager) KillPeer(ctx context.Context, addr string) (bool, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	conn, ok := m.cachePeers[addr]
	if !ok {
		return false, fmt.Errorf("%s,that does not exist", addr)
	}
	client := pb.NewGroupCacheClient(conn)
	r, err := client.Kill(ctx, &pb.KillRequest{Reason: []byte("kill peer")})
	if err != nil {
		log.Printf("could not greet: %v\n", err)
		return false, err
	}
	// 打印服务的返回的消息
	log.Printf("Greeting: %v", r.Status)
	if r.Status {
		conn.Close()
		delete(m.cachePeers, addr)
		return true, nil
	}
	return false, fmt.Errorf("kill error")
}

func (m *Manager) GetAllCache(ctx context.Context, addr string) []replacement.Entry {
	m.lock.RLock()
	defer m.lock.RUnlock()
	conn, ok := m.cachePeers[addr]
	if !ok {
		return nil
	}
	client := pb.NewGroupCacheClient(conn)
	r, err := client.GetAllCache(ctx, &pb.GetAllCacheRequest{})
	if err != nil {
		log.Printf("could not greet: %v\n", err)
		return nil
	}
	// 打印服务的返回的消息
	log.Println("Greeting", r.Size)
	res := make([]replacement.Entry, 0)
	for _, v := range r.Entry {
		l := strings.Split(string(v), ",")
		res = append(res, replacement.Entry{
			Key:   l[0],
			Value: cache.ByteView(l[1]),
		})
	}
	return res
}

func (m *Manager) GetInfo(ctx context.Context, addr string) map[string]interface{} {
	m.lock.RLock()
	defer m.lock.RUnlock()
	conn, ok := m.cachePeers[addr]
	if !ok {
		return nil
	}
	client := pb.NewGroupCacheClient(conn)
	r, err := client.Info(ctx, &pb.InfoRequest{})
	if err != nil {
		log.Printf("could not greet: %v\n", err)
		return nil
	}
	// 打印服务的返回的消息
	log.Println("Greeting")
	res := make(map[string]interface{})
	res["name"] = r.Name
	res["address"] = r.Address
	res["replacement"] = r.Replacement
	res["usedBytes"] = r.UsedBytes
	res["cacheBytes"] = r.CacheBytes
	return res
}
