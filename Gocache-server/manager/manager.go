package manager

import (
	"fmt"
	"log"
	"sync"

	"github.com/meguriri/GoCache/server/cache"
	"github.com/meguriri/GoCache/server/callback"
	"github.com/meguriri/GoCache/server/consistenthash"
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

func (m *Manager) storage(key string) (cache.ByteView, error) {
	log.Println("[Search call back] ")
	value, err := m.callback.Get(key)
	if err != nil {
		return cache.ByteView{}, fmt.Errorf("[Search call back] no local storage")
	}
	m.localCache.Add(key, cache.ByteView(value))
	return cache.ByteView(value), nil
}
