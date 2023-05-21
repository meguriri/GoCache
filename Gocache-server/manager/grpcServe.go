package manager

import (
	"context"
	"fmt"
	"log"

	"github.com/meguriri/GoCache/server/cache"
	pb "github.com/meguriri/GoCache/server/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

const (
	HEALTHCHECK_SERVICE = "grpc.health.v1.Health"
	UNKNOWN             = 0
	SERVING             = 1
	NOT_SERVING         = 2
)

// 生成一个新的peer
func (m *Manager) Connect(name, addr string, cacheBytes int64) bool {

	//加写锁
	m.lock.Lock()
	defer m.lock.Unlock()

	token := Token{
		Secret: "20001019",
	}

	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(&token),
		grpc.WithDefaultServiceConfig(
			fmt.Sprintf(`{"HealthCheckConfig": {"ServiceName": "%s"}}`,
				HEALTHCHECK_SERVICE),
		),
	)
	if err != nil {
		log.Printf("[Connect] did not connect: %v\n", err)
		return false
	}

	// 初始化Greeter服务客户端
	client := pb.NewGoCacheClient(conn)

	// 初始化上下文，设置请求超时时间为1秒
	ctx := context.Background()
	// 调用Get接口，发送一条消息
	r, err := client.Connect(ctx, &pb.ConnectRequest{Name: name, Address: addr, MaxBytes: cacheBytes})
	if err != nil {
		log.Printf("[Connect] could not greet: %v\n", err)
		return false
	}

	//生成的新group存入groups映射表
	if r.Code == 200 {
		// 打印服务的返回的消息
		log.Printf("[Connect] Greeting: %d", r.Code)
		//添加到一致性哈希中
		m.hash.Add(addr)
		//添加到节点映射表中
		m.cachePeers[addr] = conn
		//重新分配缓存
		entry := r.Entry
		m.Snapshot(entry)
	} else {
		log.Printf("[Connect] grpc connect err code: %d", r.Code)
		return false
	}
	return true
}

func (m *Manager) Get(ctx context.Context, key string) ([]byte, error) {
	//use local cache
	if len(m.cachePeers) == 0 {
		if v, ok := m.localCache.Get(key); ok {
			log.Println("[CachePeer Get] hit")
			return v.(cache.ByteView).ToByte(), nil
		}
		return nil, fmt.Errorf("no cache")
	}

	//grpc
	m.lock.RLock()
	addr := m.hash.Get(key)
	conn, ok := m.cachePeers[addr]
	m.lock.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%s,that does not exist", addr)
	}
	client := pb.NewGoCacheClient(conn)

	// 调用Get接口，发送一条消息
	r, err := client.Get(ctx, &pb.GetRequest{Key: key})

	if err != nil {

		m.lock.RLock()
		oldAddr := m.hash.GetOldPeer(key)
		oldconn, ok := m.cachePeers[oldAddr]
		m.lock.RUnlock()
		if !ok {
			return nil, fmt.Errorf("%s,that does not exist", oldAddr)
		}

		oldClient := pb.NewGoCacheClient(oldconn)
		// 调用Get接口，发送一条消息
		oldR, err := oldClient.Get(ctx, &pb.GetRequest{Key: key})
		if err != nil {
			log.Printf("[Get] could not greet: %v\n", err)
			return nil, err
		}

		delClient := pb.NewGoCacheClient(oldconn)
		dr, err := delClient.Del(ctx, &pb.DelRequest{Key: key})
		if err != nil {
			log.Printf("[Get] del could not greet: %v\n", err)
			return nil, err
		}
		log.Printf("[Get] del Greeting %v: %s\n", dr.Status, string(oldR.Value))

		sr, err := client.Set(ctx, &pb.SetRequest{Key: key, Value: oldR.Value})
		if err != nil {
			log.Printf("[Get] set could not greet: %v\n", err)
			return nil, err
		}
		log.Printf("[Get] set Greeting %v: %s\n", sr.Status, string(oldR.Value))
		return oldR.Value, nil
	}
	// 打印服务的返回的消息
	log.Printf("[Get] Greeting: %s", r.Value)
	return r.Value, nil
}

func (m *Manager) Set(ctx context.Context, key string, value []byte) bool {
	if len(m.cachePeers) == 0 {
		m.localCache.Add(key, cache.ByteView(value))
		m.ModifyCnt++
		return true
	}

	// m.lock.RLock()
	// oldAddr := m.hash.GetOldPeer(key)
	// oldconn, ok := m.cachePeers[oldAddr]
	// m.lock.RUnlock()
	// if !ok {
	// 	return false
	// }
	// oldClient := pb.NewGoCacheClient(oldconn)
	// // 调用Get接口，发送一条消息
	// oldR, err := oldClient.Get(ctx, &pb.GetRequest{Key: key})
	// if err == nil {
	// 	log.Println("[Set] get old value not nil", string(oldR.Value))
	// 	oldClient.Del(ctx, &pb.DelRequest{Key: key})
	// }

	m.lock.RLock()
	addr := m.hash.Get(key)
	conn, ok := m.cachePeers[addr]
	m.lock.RUnlock()
	if !ok {
		return false
	}

	client := pb.NewGoCacheClient(conn)
	// 调用Get接口，发送一条消息
	r, err := client.Set(ctx, &pb.SetRequest{Key: key, Value: value})
	if err != nil {
		log.Printf("[set] could not greet: %v\n", err)
		return false
	}
	// 打印服务的返回的消息
	log.Printf("[set] Greeting: %v", r.Status)
	return r.Status
}

func (m *Manager) Del(ctx context.Context, key string) bool {
	if len(m.cachePeers) == 0 {
		ok := m.localCache.Delete(key)
		if ok {
			m.ModifyCnt++
		}
		return ok
	}

	m.lock.RLock()
	addr := m.hash.Get(key)
	conn, ok := m.cachePeers[addr]
	m.lock.RUnlock()
	if !ok {
		return false
	}

	client := pb.NewGoCacheClient(conn)

	// 调用Get接口，发送一条消息
	r, err := client.Del(ctx, &pb.DelRequest{Key: key})
	if err != nil {
		log.Printf("[Del] could not greet: %v\n", err)
		return false
	}
	// 打印服务的返回的消息
	log.Printf("[Del] Greeting: %v", r.Status)
	return r.Status
}

// 删除一个节点
func (m *Manager) Kill(ctx context.Context, addr string) (bool, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if len(m.cachePeers) == 1 {
		return false, fmt.Errorf("there is only one peer,kill false")
	}
	conn, ok := m.cachePeers[addr]
	if !ok {
		return false, fmt.Errorf("%s ,that does not exist", addr)
	}
	client := pb.NewGoCacheClient(conn)
	r, err := client.Kill(ctx, &pb.KillRequest{Reason: []byte("kill peer")})
	if err != nil {
		log.Printf("[Kill] could not greet: %v\n", err)
		return false, err
	}
	// 打印服务的返回的消息
	log.Printf("[Kill] Greeting: %v", r.Status)
	if r.Status {
		conn.Close()
		m.RefreshCache(ctx, addr, r.Entry)
		return true, nil
	}
	return false, fmt.Errorf("kill error")
}

func (m *Manager) GetAllCache(ctx context.Context, addr string) [][]byte {
	m.lock.RLock()
	defer m.lock.RUnlock()
	conn, ok := m.cachePeers[addr]
	if !ok {
		return nil
	}
	client := pb.NewGoCacheClient(conn)
	r, err := client.GetAllCache(ctx, &pb.GetAllCacheRequest{})
	if err != nil {
		log.Printf("[GetAllCache] could not greet: %v\n", err)
		return nil
	}
	// 打印服务的返回的消息
	log.Println("[GetAllCache] Greeting", r.Size)
	return r.Entry
}

func (m *Manager) Info(ctx context.Context, addr string) map[string]interface{} {
	m.lock.RLock()
	defer m.lock.RUnlock()
	conn, ok := m.cachePeers[addr]
	if !ok {
		return nil
	}
	client := pb.NewGoCacheClient(conn)
	r, err := client.Info(ctx, &pb.InfoRequest{})
	if err != nil {
		log.Printf("[Info] could not greet: %v\n", err)
		return nil
	}
	// 打印服务的返回的消息
	log.Println("[Info] Greeting")
	res := make(map[string]interface{})
	res["name"] = r.Name
	res["address"] = r.Address
	res["replacement"] = r.Replacement
	res["usedBytes"] = r.UsedBytes
	res["cacheBytes"] = r.CacheBytes
	return res
}

func (m *Manager) Check(ctx context.Context) {
	for addr, conn := range m.cachePeers {
		client := healthpb.NewHealthClient(conn)
		// 调用Get接口，发送一条消息
		r, err := client.Check(ctx, &healthpb.HealthCheckRequest{Service: HEALTHCHECK_SERVICE})
		if err != nil {
			log.Printf("[Check] Service %s is dead", addr)
			m.lock.Lock()
			conn.Close()
			delete(m.cachePeers, addr)
			m.lock.Unlock()
			continue
		}
		// 打印服务的返回的消息
		log.Printf("[Check] Service %s is %s", addr, r.Status)
	}
}
