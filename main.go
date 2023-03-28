package main

import (
	"fmt"
	"log"

	"github.com/meguriri/GoCache/cache"
	"github.com/meguriri/GoCache/callback"
	"github.com/meguriri/GoCache/communicated/http"
	"github.com/meguriri/GoCache/config"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {

	if err := config.Configinit(); err != nil {
		log.Println("config err:", err)
	}

	cache.NewGroup("scores", 2<<10, callback.CallBackFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "localhost:9999"
	r := http.RouterInit()
	r.Run(addr)
}
