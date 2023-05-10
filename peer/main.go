package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/meguriri/GoCache/peer/cache"
	"github.com/meguriri/GoCache/peer/callback"
	"github.com/meguriri/GoCache/peer/config"
)

func main() {
	cb := func(key string) ([]byte, error) {
		if value, ok := cache.DB[key]; ok {
			return []byte(value), nil
		}
		return []byte{}, fmt.Errorf("[call back] no local storage")
	}
	if err := config.Configinit(); err != nil {
		log.Println("[config init] err:", err.Error())
		return
	}
	peer := cache.NewPeer(callback.CallBackFunc(cb))
	// peer.ReadLocalStorage()
	ctx, cancel := context.WithCancel(context.Background())
	go peer.Save(ctx)
	go peer.Listen(ctx)
	for {
		select {
		case <-peer.KillSignal:
			cancel()
			time.Sleep(time.Second * 3)
			log.Println("cache is stop:", time.Now())
			return
		}
	}
}
