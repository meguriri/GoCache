package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/meguriri/GoCache/server/cache"
	"github.com/meguriri/GoCache/server/callback"
	"github.com/meguriri/GoCache/server/config"
	"github.com/meguriri/GoCache/server/manager"
)

func CMD(ctx context.Context, m *manager.Manager) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(manager.ManagerIP+":"+manager.ManagerPort, ">")
		res, _, _ := reader.ReadLine()
		li := strings.Split(string(res), " ")
		if (li[0] == "set" || li[0] == "SET") && len(li) == 3 {
			if ok := m.Set(ctx, li[1], []byte(li[2])); !ok {
				fmt.Println("set error")
			} else {
				fmt.Println("OK")
			}
		} else if (li[0] == "get" || li[0] == "GET") && len(li) == 2 {
			res, err := m.Get(ctx, li[1])
			if err != nil {
				fmt.Println("(nil)")
			} else {
				fmt.Println("\"", string(res), "\"")
			}
		} else if (li[0] == "del" || li[0] == "DEL") && len(li) == 2 {
			res := m.Del(ctx, li[1])
			if !res {
				fmt.Println("(integer) 0")
			} else {
				fmt.Println("(integer) 1")
			}
		} else if li[0] == "exit" {
			fmt.Println("bye!")
			return
		} else if li[0] == "kill" && len(li) == 2 {
			if ok, err := m.KillPeer(ctx, li[1]); ok {
				fmt.Println(li[1], "is logout")
			} else {
				fmt.Println(li[1], "logout err:", err)
			}
		} else if li[0] == "connect" && len(li) == 4 {
			bytes, _ := strconv.Atoi(li[3])
			if ok := m.Connect(li[1], li[2], int64(bytes)); ok {
				fmt.Println(li[2], "is connected")
			} else {
				fmt.Println(li[2], "connect err:")
			}
		} else if (li[0] == "getall" || li[0] == "GETALL") && len(li) == 2 {
			res := m.GetAllCache(ctx, li[1])
			fmt.Println("res:", res)
		} else if (li[0] == "info" || li[0] == "INFO") && len(li) == 2 {
			res := m.GetInfo(ctx, li[1])
			fmt.Println("res:", res)
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

	manager := manager.NewManager(callback.CallBackFunc(cb))
	manager.Connect("peer1", "127.0.0.1:8081", 1024*1024)
	// manager.Connect("peer2", "127.0.0.1:8082", 1024*1024)
	// manager.Connect("peer3", "127.0.0.1:8083", 1024*1024)

	ctx := context.Background()
	CMD(ctx, manager)

}
