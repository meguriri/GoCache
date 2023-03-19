package replacement

import (
	"log"

	"github.com/meguriri/GoCache/data"
	"github.com/meguriri/GoCache/replacement/fifo"
	"github.com/meguriri/GoCache/replacement/lfu"
	"github.com/meguriri/GoCache/replacement/lru"
)

func NewCache(t string) data.Cache {
	switch t {
	case "FIFO":
		log.Println("use FIFO")
		return fifo.New(nil)
	case "LFU":
		log.Println("use LFU")
		return lfu.New(nil)
	case "LRU":
		log.Println("use LRU")
		return lru.New(nil)
	}
	return nil
}
