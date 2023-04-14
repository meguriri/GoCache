package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/meguriri/GoCache/cache"
	"github.com/meguriri/GoCache/callback"
	"github.com/meguriri/GoCache/config"
	"github.com/meguriri/GoCache/manager"
	"github.com/meguriri/GoCache/proto"
)

func main() {
	if err := config.Configinit(); err != nil {
		log.Fatal("config error:", err)
	}
	manager := manager.NewManager()
	cb := func(key string) ([]byte, error) {
		if value, ok := cache.DB[key]; ok {
			return []byte(value), nil
		}
		return []byte{}, fmt.Errorf("[call back] no local storage")
	}
	manager.NewPeer("http://127.0.0.1:8081", 1024, callback.CallBackFunc(cb))
	manager.NewPeer("http://127.0.0.1:8082", 1024, callback.CallBackFunc(cb))
	manager.NewPeer("http://127.0.0.1:8083", 1024, callback.CallBackFunc(cb))

	//manager.Connect()
	//time.Sleep(time.Second * 5)
	ctx := context.Background()
	for i := 0; i <= 10; i++ {
		req := &proto.CacheRequest{Group: strconv.Itoa(i), Key: strconv.Itoa(i)}
		res, err := manager.Get(ctx, req)
		if err != nil {
			log.Println(i, " get error1: ", err)
		}
		fmt.Println(i, " answer1: ", string(res.GetValue()))
		res, err = manager.Get(ctx, req)
		if err != nil {
			log.Println(i, " get error2: ", err)
		}
		fmt.Println(i, " answer2: ", string(res.GetValue()))
	}
}
