package cache

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/meguriri/GoCache/peer/proto"
	"github.com/meguriri/GoCache/peer/replacement"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

const HEALTHCHECK_SERVICE = "grpc.health.v1.Health"

func Check(ctx context.Context) error {
	//从上下文中获取元数据
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "获取Token失败")
	}
	var Secret string
	if value, ok := md["secret"]; ok {
		Secret = value[0]
	}
	if Secret != "20001019" {
		return status.Errorf(codes.Unauthenticated, "Token无效:secret=%s", Secret)
	}
	return nil
}

func (c *CachePeer) Listen(ctx context.Context) {
	log.Println("[Listen] grpc server listen on:", c.Addr)
	lis, err := net.Listen("tcp", c.Addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return
	}
	// 实例化grpc服务端
	c.Server = grpc.NewServer()

	//health
	healthserver := health.NewServer()
	healthserver.SetServingStatus(HEALTHCHECK_SERVICE, healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(c.Server, healthserver)

	// 注册Greeter服务
	pb.RegisterGoCacheServer(c.Server, c)
	// 往grpc服务端注册反射服务
	reflection.Register(c.Server)
	// 启动grpc服务
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				log.Println("[listen] done")
				c.Server.Stop()
				return
			default:
				{
				}
			}
		}
	}(ctx)
	if err := c.Server.Serve(lis); err != nil {
		log.Fatalf("[listen] failed to serve: %v", err)
	}
}

func (c *CachePeer) Connect(ctx context.Context, in *pb.ConnectRequest) (*pb.ConnectResponse, error) {

	if err := Check(ctx); err != nil {
		return nil, err
	}

	c.Lock.Lock()
	defer c.Lock.Unlock()
	if in.Address != c.Addr {
		return &pb.ConnectResponse{Code: 404, Entry: nil}, fmt.Errorf("[Invalid request]server address is not %s", in.Address)
	}
	log.Println("[CachePeer Connect]", in.Address, "success!")
	c.Name, c.CacheBytes = in.Name, in.MaxBytes
	res := c.ReadLocalStorage()
	return &pb.ConnectResponse{Code: 200, Entry: res}, nil
}

func (c *CachePeer) Get(ctx context.Context, in *pb.GetRequest) (out *pb.GetResponse, err error) {

	if err := Check(ctx); err != nil {
		return nil, err
	}

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
		return out, fmt.Errorf("[CachePeer Get] %s no local cache area", c.Addr)
	}

	//从底层缓存区获取缓存值
	if v, ok := c.Manager.Get(key); ok {
		log.Printf("[CachePeer Get] hit key: %s value: %s\n", key, string(v.(ByteView).ToByte()))
		return &pb.GetResponse{Value: v.(ByteView).ToByte()}, nil
	}
	log.Println("[CachePeer Get] Cache Miss", key)
	return out, fmt.Errorf("[CachePeer Get] %s Cache Miss", c.Addr)
}

func (c *CachePeer) Set(ctx context.Context, in *pb.SetRequest) (out *pb.SetResponse, err error) {

	if err := Check(ctx); err != nil {
		return nil, err
	}

	c.Lock.Lock()
	defer c.Lock.Unlock()
	key, value := in.Key, in.Value
	c.Manager.Add(key, ByteView(value))
	log.Printf("[CachePeer Set] key: %s value: %s\n", key, string(value))
	c.ModifyCnt++
	return &pb.SetResponse{Status: true}, nil
}

func (c *CachePeer) Del(ctx context.Context, in *pb.DelRequest) (out *pb.DelResponse, err error) {

	if err := Check(ctx); err != nil {
		return nil, err
	}

	c.Lock.Lock()
	defer c.Lock.Unlock()
	key := in.Key
	ok := c.Manager.Delete(key)
	if ok {
		c.ModifyCnt++
		log.Printf("[CachePeer Del] key: %s\n", key)
		return &pb.DelResponse{Status: ok}, nil
	}
	return &pb.DelResponse{Status: ok}, fmt.Errorf("del error")
}

func (c *CachePeer) Kill(ctx context.Context, in *pb.KillRequest) (*pb.KillResponse, error) {

	if err := Check(ctx); err != nil {
		return nil, err
	}

	c.Lock.Lock()
	defer c.Lock.Unlock()
	log.Println("[CachePeer kill] reaseon: ", string(in.Reason))
	entry := c.Manager.GetAll()
	c.ticker.Stop()
	c.Snapshot(nil)
	c.KillSignal <- struct{}{}
	return &pb.KillResponse{Status: true, Entry: entry}, nil
}

func (c *CachePeer) GetAllCache(ctx context.Context, in *pb.GetAllCacheRequest) (*pb.GetAllCacheResponse, error) {

	if err := Check(ctx); err != nil {
		return nil, err
	}
	c.Lock.RLock()
	defer c.Lock.RUnlock()
	res := c.Manager.GetAll()
	return &pb.GetAllCacheResponse{Size: int64(len(res)), Entry: res}, nil
}

func (c *CachePeer) Info(ctx context.Context, in *pb.InfoRequest) (*pb.InfoResponse, error) {

	if err := Check(ctx); err != nil {
		return nil, err
	}
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
