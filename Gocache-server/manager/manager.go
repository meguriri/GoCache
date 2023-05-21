package manager

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/meguriri/GoCache/server/cache"
	"github.com/meguriri/GoCache/server/config"
	"github.com/meguriri/GoCache/server/consistenthash"
	"github.com/meguriri/GoCache/server/replacement"
	"github.com/meguriri/GoCache/server/replacement/manager"
	"google.golang.org/grpc"
)

var (
	ManagerIP       string
	ManagerPort     string
	DefaultReplicas int //虚拟节点个数
	SaveSeconds     int
	SaveModify      int
	wg              sync.WaitGroup
)

type Manager struct {
	addr            string
	lock            sync.RWMutex                //peers并发访问安全的读写锁
	cachePeers      map[string]*grpc.ClientConn //peer名与grpc conn的映射表
	localCache      replacement.CacheManager
	localcacheBytes int64
	ticker          *time.Ticker
	ModifyCnt       int
	hash            consistenthash.Map
}

func NewManager() *Manager {
	return &Manager{
		addr:            ManagerIP + ":" + ManagerPort,
		lock:            sync.RWMutex{},
		cachePeers:      make(map[string]*grpc.ClientConn),
		localCache:      manager.NewCache(replacement.ReplacementPolicy),
		localcacheBytes: replacement.MaxBytes,
		hash:            *consistenthash.New(DefaultReplicas, nil),
	}
}

func StartServer() error {
	config, err := config.ConfigInit()
	if err != nil {
		return err
	}
	replacement.ReplacementPolicy = config.GetString("replacement.policy")
	replacement.MaxBytes = config.GetInt64("replacement.max-bytes")
	DefaultReplicas = config.GetInt("defaultReplicas")
	SaveSeconds = config.GetInt("save.seconds")
	SaveModify = config.GetInt("save.modify")
	ManagerIP = config.GetString("server.ip")
	ManagerPort = config.GetString("server.port")

	manager := NewManager()
	peerList := config.Get("peers").([]interface{})
	for _, peer := range peerList {
		manager.Connect(peer.(map[string]interface{})["name"].(string),
			peer.(map[string]interface{})["address"].(string),
			int64(peer.(map[string]interface{})["cache_bytes"].(int)),
		)
	}
	ctx, cancel := context.WithCancel(context.Background())
	manager.Init(ctx)
	go manager.TCPServe(ctx, &wg)
	go manager.HeartBeat(ctx, &wg)

	wg.Add(2)
	wg.Wait()
	cancel()
	return nil
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

func (m *Manager) HeartBeat(ctx context.Context, wg *sync.WaitGroup) {
	ticker := time.NewTicker(time.Second * time.Duration(10))
	for range ticker.C {
		m.Check(ctx)
	}
	for {
		select {
		case <-ticker.C:
			m.Check(ctx)
		case <-ctx.Done():
			fmt.Println("heartBeat over")
			return
		}
	}

}

func (m *Manager) RefreshCache(ctx context.Context, addr string, entry [][]byte) {
	//删除节点映射表
	delete(m.cachePeers, addr)
	//刷新一致性哈希
	m.hash.Refresh()
	for addr := range m.cachePeers {
		m.hash.Add(addr)
	}
	//重新分配缓存
	kv := ToEntry(entry)
	for k, v := range kv {
		m.Set(ctx, k, cache.ByteView(v))
	}
}

func ToEntry(data [][]byte) map[string][]byte {
	m := make(map[string][]byte)
	for _, b := range data {
		l := strings.Split(string(b), ",")
		log.Println(l[0], l[1])
		m[l[0]] = []byte(l[1])[:len(l[1])-1]
	}
	return m
}

func (m *Manager) Save(ctx context.Context) {
	m.ticker = time.NewTicker(time.Second * time.Duration(SaveSeconds))
	for {
		select {
		case <-m.ticker.C:
			if m.ModifyCnt >= SaveModify {
				m.ModifyCnt = 0
				res := m.localCache.GetAll()
				m.Snapshot(res)
			}
		case <-ctx.Done():
			log.Println("save done")
			return
		}
	}
}

func (m *Manager) Snapshot(res [][]byte) {
	file, err := os.OpenFile("save.gdb", os.O_APPEND, 0666)
	if err != nil {
		log.Println("[Snapshot] open save.gdb err", err)
		return
	}
	defer file.Close()
	for _, v := range res {
		n, err := file.Write(v)
		if err != nil {
			log.Println("[Snapshot] file write err", err.Error())
			continue
		}
		file.Write([]byte("\r\n"))
		log.Printf("[Snapshot] write %d bytes\n", n)
	}
}

func (m *Manager) Init(ctx context.Context) {
	if len(m.cachePeers) == 0 {
		log.Println("use readLocalStorge")
		m.readLocalStorage()
		return
	}
	log.Println("use allocationCache")
	m.allocationCache(ctx)
}

// 将save.gdb中的内容分配到不同节点当中
func (m *Manager) allocationCache(ctx context.Context) {
	bytes, err := os.ReadFile("save.gdb")
	if err != nil {
		log.Println("[ReadLocalStorage] read file err", err.Error())
	}
	log.Println("[allocationCache] readall", string(bytes))
	l := strings.Split(string(bytes), "\r\n")
	for _, v := range l {
		if v != "" {
			entry := strings.Split(v, ",")
			m.Set(ctx, entry[0], cache.ByteView(entry[1]))
		}
	}
	os.OpenFile("save.gdb", os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0766)
}

func (m *Manager) readLocalStorage() {
	bytes, err := os.ReadFile("save.gdb")
	if err != nil {
		log.Println("[ReadLocalStorage] read file err", err.Error())
	}
	l := strings.Split(string(bytes), "\r\n")
	for _, v := range l {
		if v != "" {
			entry := strings.Split(v, ",")
			m.localCache.Add(entry[0], cache.ByteView(entry[1]))
		}
	}
}
