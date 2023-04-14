package manager

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/meguriri/GoCache/cache"
	"github.com/meguriri/GoCache/callback"
	"github.com/meguriri/GoCache/consistenthash"
	pb "github.com/meguriri/GoCache/proto"
	"github.com/meguriri/GoCache/replacement"
	"github.com/meguriri/GoCache/replacement/manager"
	"google.golang.org/protobuf/proto"
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

	go http.ListenAndServe(addr[7:], peer)

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

func (m *Manager) Connect() {
	for k, v := range m.cachePeers {
		fmt.Println("[Connect]", k)
		go http.ListenAndServe(k[7:], v)
	}
}

func (m *Manager) Get(ctx context.Context, req *pb.CacheRequest) (*pb.CacheResponse, error) {
	_, key := req.GetGroup(), req.GetKey()
	res := &pb.CacheResponse{}
	//use local cache
	if len(m.cachePeers) == 0 {
		if v, ok := m.localCache.Get(key); ok {
			res.Value = v.(cache.ByteView).GetByte()
			return res, nil
		}
		return res, fmt.Errorf("no cache")
	}

	addr := m.hash.Get(key)
	//GET请求URL
	resp, err := http.Get(addr + "/" + key)
	if err != nil {
		log.Println("err: ", err)
		return res, fmt.Errorf("server returned: %v", resp.Status)
	}
	defer resp.Body.Close()

	//resp状态不为 200
	if resp.StatusCode != http.StatusOK {
		return res, fmt.Errorf("server returned: %v", resp.Status)
	}

	//读取response.body
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, fmt.Errorf("reading response body: %v", err)
	}

	//proto反序列化
	if err = proto.Unmarshal(bytes, res); err != nil {
		return res, fmt.Errorf("decoding response body: %v", err)
	}

	fmt.Println("[http Get]", string(res.GetValue()))
	return res, nil
}
