package manager

import (
	"context"
	"fmt"
	"log"
	"time"

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
		log.Printf("did not connect: %v\n", err)
		return false
	}

	// 初始化Greeter服务客户端
	client := pb.NewGoCacheClient(conn)

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
	client := pb.NewGoCacheClient(conn)
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
	client := pb.NewGoCacheClient(conn)

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
	client := pb.NewGoCacheClient(conn)

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

func (m *Manager) Kill(ctx context.Context, addr string) (bool, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	conn, ok := m.cachePeers[addr]
	if !ok {
		return false, fmt.Errorf("%s,that does not exist", addr)
	}
	client := pb.NewGoCacheClient(conn)
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
		log.Printf("could not greet: %v\n", err)
		return nil
	}
	// 打印服务的返回的消息
	log.Println("Greeting", r.Size)
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

func (m *Manager) HeartBeat(ctx context.Context) {
	ticker := time.NewTicker(time.Second * time.Duration(5))
	for range ticker.C {
		m.Check(ctx)
	}
	fmt.Println("heartBeat over")
}

func (m *Manager) Check(ctx context.Context) {
	for addr, conn := range m.cachePeers {
		client := healthpb.NewHealthClient(conn)

		// 调用Get接口，发送一条消息
		r, err := client.Check(ctx, &healthpb.HealthCheckRequest{Service: HEALTHCHECK_SERVICE})
		if err != nil {

			log.Printf("Service %s is dead", addr)
			// m.lock.Lock()
			// conn.Close()
			// delete(m.cachePeers, addr)
			// m.lock.Unlock()
			continue
		}
		// 打印服务的返回的消息
		log.Printf("Service %s is %s", addr, r.Status)
	}
}
