package main

import (
	"fmt"
	"log"

	"github.com/meguriri/GoCache/server/cache"
	"github.com/meguriri/GoCache/server/callback"
	"github.com/meguriri/GoCache/server/config"
	"github.com/meguriri/GoCache/server/manager"
)

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
	//manager.Connect("peer1", "127.0.0.1:8081", 1024*1024)

	manager.Serve()
}
