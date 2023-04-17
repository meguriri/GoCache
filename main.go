package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/meguriri/GoCache/cache"
	"github.com/meguriri/GoCache/callback"
	"github.com/meguriri/GoCache/config"
	"github.com/meguriri/GoCache/manager"
	"github.com/meguriri/GoCache/proto"
	pb "github.com/meguriri/GoCache/proto"
)

func CMD(ctx context.Context, m *manager.Manager) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(m.GetAddr(), ">")
		res, _, _ := reader.ReadLine()
		li := strings.Split(string(res), " ")
		if (li[0] == "set" || li[0] == "SET") && len(li) == 3 {
			if ok, err := m.Set(&pb.SetRequest{Key: li[1], Value: cache.ByteView([]byte(li[2]))}); !ok.Status {
				fmt.Println("err: ", err)
			} else {
				fmt.Println("OK")
			}
		} else if (li[0] == "get" || li[0] == "GET") && len(li) == 2 {
			res, err := m.Get(ctx, &proto.GetRequest{Key: li[1]})
			if err != nil {
				fmt.Println("(nil)")
			} else {
				fmt.Println("\"", string(res.Value), "\"")
			}
		} else if (li[0] == "del" || li[0] == "DEL") && len(li) == 2 {
			res, _ := m.Del(&pb.DelRequest{Key: li[1]})
			if !res.Status {
				fmt.Println("(integer) 0")
			} else {
				fmt.Println("(integer) 1")
			}
		} else if li[0] == "exit" {
			fmt.Println("bye!")
			return
		}
	}
}

func main() {
	if err := config.Configinit(); err != nil {
		log.Fatal("config error:", err)
	}

	cb := func(key string) ([]byte, error) {
		if value, ok := cache.DB[key]; ok {
			return []byte(value), nil
		}
		return []byte{}, fmt.Errorf("[call back] no local storage")
	}

	manager := manager.NewManager("127.0.0.1:8080", 1024*1024, callback.CallBackFunc(cb))
	manager.NewPeer("peer1", "127.0.0.1:8081", 1024, callback.CallBackFunc(cb))
	manager.NewPeer("peer2", "127.0.0.1:8082", 1024, callback.CallBackFunc(cb))
	manager.NewPeer("peer3", "127.0.0.1:8083", 1024, callback.CallBackFunc(cb))

	ctx := context.Background()
	// for i := 0; i <= 10; i++ {
	// 	req := &proto.CacheRequest{Key: strconv.Itoa(i)}
	// 	res, err := manager.Get(ctx, req)
	// 	if err != nil {
	// 		log.Println(i, " get error1: ", err)
	// 	} else {
	// 		fmt.Println(i, " answer1: ", string(res.GetValue()))
	// 	}
	// 	res, err = manager.Get(ctx, req)
	// 	if err != nil {
	// 		log.Println(i, " get error2: ", err)
	// 	} else {
	// 		fmt.Println(i, " answer2: ", string(res.GetValue()))
	// 	}
	// }
	CMD(ctx, manager)

}
