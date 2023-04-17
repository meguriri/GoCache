package cache

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/meguriri/GoCache/callback"
	pb "github.com/meguriri/GoCache/proto"
	"github.com/meguriri/GoCache/replacement"
	"google.golang.org/grpc"
)

type CachePeer struct {
	pb.UnimplementedGroupCacheServer
	Server       *grpc.Server
	Name         string
	Addr         string                   //cache的地址
	CallBackFunc callback.CallBack        //缓存未命中的回调函数
	Lock         sync.Mutex               //互斥锁
	Manager      replacement.CacheManager //cache 底层存储和淘汰算法
	CacheBytes   int64                    //缓存字节最大值
}

func (c *CachePeer) Get(ctx context.Context, in *pb.GetRequest) (out *pb.GetResponse, err error) {
	key := in.GetKey()
	//验证key合法性
	if key == "" {
		return out, fmt.Errorf("key is required")
	}

	//从本地节点获取缓存值
	c.Lock.Lock()
	defer c.Lock.Unlock()

	//cache未绑定底层缓存区
	if c.Manager == nil {
		return out, fmt.Errorf("[CachePeers Get] %s 未绑定底层缓存区", c.Addr)
	}

	//从底层缓存区获取缓存值
	if v, ok := c.Manager.Get(key); ok {
		log.Println("[Cache] hit")
		return &pb.GetResponse{Value: v.(ByteView).GetByte()}, nil
	} else if c.CallBackFunc != nil {
		view, err := c.storage(key)
		if err != nil {
			return out, err
		}
		return &pb.GetResponse{Value: view.GetByte()}, nil
	}
	return out, fmt.Errorf("[CachePeers Get] %s 缓存未命中", c.Addr)
}

func (c *CachePeer) Set(ctx context.Context, in *pb.SetRequest) (out *pb.SetResponse, err error) {
	key, value := in.Key, in.Value
	c.Manager.Add(key, ByteView(value))
	return &pb.SetResponse{Status: true}, nil
}

func (c *CachePeer) Del(ctx context.Context, in *pb.DelRequest) (out *pb.DelResponse, err error) {
	key := in.Key
	c.Manager.Delete(key)
	return &pb.DelResponse{Status: true}, nil
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
