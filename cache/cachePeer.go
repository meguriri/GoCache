package cache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/meguriri/GoCache/callback"
	pb "github.com/meguriri/GoCache/proto"
	"github.com/meguriri/GoCache/replacement"
	"google.golang.org/protobuf/proto"
)

type CachePeer struct {
	Addr         string                   //cache的地址
	CallBackFunc callback.CallBack        //缓存未命中的回调函数
	Lock         sync.Mutex               //互斥锁
	Manager      replacement.CacheManager //cache 底层存储和淘汰算法
	CacheBytes   int64                    //缓存字节最大值
}

// 获取缓存值
func (c *CachePeer) Get(key string) (ByteView, error) {

	//验证key合法性
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	//从本地节点获取缓存值
	c.Lock.Lock()
	defer c.Lock.Unlock()

	//cache未绑定底层缓存区
	if c.Manager == nil {
		return ByteView{}, fmt.Errorf("[CachePeers Get] %s 未绑定底层缓存区", c.Addr)
	}

	//从底层缓存区获取缓存值
	if v, ok := c.Manager.Get(key); ok {
		log.Println("[Cache] hit")
		return v.(ByteView), nil
	} else if c.CallBackFunc != nil {
		view, err := c.storage(key)
		if err != nil {
			return ByteView{}, err
		}
		return view, nil
	}
	return ByteView{}, fmt.Errorf("[CachePeers Get] %s 缓存未命中", c.Addr)
}

func (c *CachePeer) storage(key string) (ByteView, error) {
	log.Println("[Search call back] ")
	value, err := c.CallBackFunc.Get(key)
	if err != nil {
		return ByteView{}, fmt.Errorf("[Search call back] no local storage")
	}
	c.Manager.Add(key, ByteView{b: value})
	return ByteView{b: value}, nil
}

func (c *CachePeer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("[ServeHTTP] basePath", c.Addr)
	log.Println("[ServeHTTP] request URL", r.URL.Path)
	//打印请求方法和请求路径
	log.Printf("[ServeHTTP] %s %s\n", r.Method, r.URL.Path)

	//获取节点名称和key
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	key := parts[1]
	log.Println("[ServeHTTP] key", key)

	//获取数据
	view, err := c.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("[ServeHTTP] get value ", string(view.GetByte()))

	//将缓存数据proto序列化
	body, err := proto.Marshal(&pb.CacheResponse{Value: view.GetByte()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//传输数据为二进制流
	w.Header().Set("Content-Type", "application/octet-stream")

	//传入数据(value)
	w.Write(body)
}
