package manager

import (
	"log"

	"github.com/meguriri/GoCache/server/replacement"
	"github.com/meguriri/GoCache/server/replacement/fifo"
	"github.com/meguriri/GoCache/server/replacement/lfu"
	"github.com/meguriri/GoCache/server/replacement/lru"
)

func NewCache(t string) replacement.CacheManager {
	log.Printf("[NewCache] ")
	switch t {
	case "FIFO":
		log.Printf("use FIFO\n")
		return fifo.New(nil)
	case "LFU":
		log.Printf("use LFU\n")
		return lfu.New(nil)
	case "LRU":
		log.Printf("use LRU\n")
		return lru.New(nil)
	}
	return nil
}
