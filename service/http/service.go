package http

// import (
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"strings"
// 	"sync"

// 	"github.com/meguriri/GoCache/cache"
// 	"github.com/meguriri/GoCache/communicated"
// 	"github.com/meguriri/GoCache/consistenthash"
// 	pb "github.com/meguriri/GoCache/proto"
// 	"google.golang.org/protobuf/proto"
// )

// var (
// 	DefaultbasePath string //默认节点url地址
// 	DefaultReplicas int    //虚拟节点个数
// )

// // HTTP服务
// type HTTPService struct {
// 	self       string                 //自己的IP和端口
// 	basePath   string                 //节点的url地址
// 	lock       sync.Mutex             //互斥锁
// 	peers      *consistenthash.Map    //一致性哈希环
// 	httpGetter map[string]*httpGetter //一个节点对应一个httpGetter的映射表
// }

// // 创建一个新的HTTP服务
// func NewHttpService(self, baseURL string) *HTTPService {
// 	return &HTTPService{
// 		self:     self,
// 		basePath: baseURL,
// 	}
// }

// // 日志
// func (h *HTTPService) Log(format string, v ...interface{}) {
// 	log.Printf("[Server %s] %s", h.self, fmt.Sprintf(format+"\n", v...))
// }

// // 实现Server接口
// func (h *HTTPService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	log.Println("[ServeHTTP] basePath", h.basePath)
// 	log.Println("[ServeHTTP] request URL", r.URL.Path)
// 	//baseURL无效
// 	// if !strings.HasPrefix(r.URL.Host, h.basePath) {
// 	// 	panic("HTTPPool serving unexpected path: " + r.URL.Path)
// 	// }

// 	//打印请求方法和请求路径
// 	h.Log("%s %s", r.Method, r.URL.Path)

// 	//获取节点名称和key
// 	parts := strings.Split(r.URL.Path, "/")
// 	log.Println("[ServeHTTP] parts", parts)
// 	log.Println("[ServeHTTP] parts len", len(parts))
// 	if len(parts) != 3 {
// 		http.Error(w, "bad request", http.StatusBadRequest)
// 		return
// 	}

// 	groupName, key := parts[1], parts[2]
// 	log.Println("[ServeHTTP] groupName", groupName)
// 	log.Println("[ServeHTTP] key", key)

// 	//获取该节点对应的group
// 	group := cache.GetGroup(groupName)
// 	if group == nil {
// 		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
// 		return
// 	}

// 	//获取数据
// 	view, err := group.Get(key)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	//将缓存数据proto序列化
// 	body, err := proto.Marshal(&pb.Response{Value: view.GetByte()})
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	//传输数据为二进制流
// 	w.Header().Set("Content-Type", "application/octet-stream")

// 	//传入数据(value)
// 	w.Write(body)
// }

// // 节点初始化
// func (h *HTTPService) Set(peers ...string) {

// 	//加锁
// 	h.lock.Lock()
// 	defer h.lock.Unlock()

// 	//生成一致性哈希环
// 	h.peers = consistenthash.New(DefaultReplicas, nil)

// 	//一致性哈希环上添加节点
// 	h.peers.Add(peers...)

// 	//初始化httpGetter map
// 	h.httpGetter = make(map[string]*httpGetter, len(peers))

// 	for _, peer := range peers {
// 		//每个节点的getter存入映射表中
// 		h.httpGetter[peer] = &httpGetter{baseURL: peer}
// 	}
// }

// // 根据缓存的Key获取相应远程节点的Getter
// func (h *HTTPService) PickPeer(key string) (communicated.PeerGetter, bool) {

// 	//加锁
// 	h.lock.Lock()
// 	defer h.lock.Unlock()

// 	//获取节点的地址
// 	if peer := h.peers.Get(key); peer != "" && peer != h.self {
// 		h.Log("Pick peer %s", peer)

// 		//从映射表中获取对应的Getter
// 		return h.httpGetter[peer], true
// 	}

// 	return nil, false
// }
