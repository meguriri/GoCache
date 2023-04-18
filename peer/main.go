package main

import (
	"fmt"
	"log"

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
	peer.ReadLocalStorage()
	go peer.Save()
	peer.Listen()
}
