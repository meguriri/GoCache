package main

import (
	"context"
	"log"
	"time"

	"github.com/meguriri/GoCache/peer/cache"
	"github.com/meguriri/GoCache/peer/config"
)

func main() {
	if err := config.Configinit(); err != nil {
		log.Println("[config init] err:", err.Error())
		return
	}
	peer := cache.NewPeer()
	// peer.ReadLocalStorage()
	ctx, cancel := context.WithCancel(context.Background())
	//go peer.Save(ctx)
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
