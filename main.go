package main

import (
	"log"

	"github.com/meguriri/GoCache/config"
	"github.com/meguriri/GoCache/data"
	"github.com/meguriri/GoCache/replacement"
)

type String string

func (s String) Len() int {
	return len(s)
}

func main() {

	if err := config.Configinit(); err != nil {
		log.Println("config err:", err)
	}

	cache := replacement.NewCache(data.ReplacementPolicy)
	cache.Add("1", String("1234"))
	cache.GetAll()

	cache.Add("2", String("fuck"))
	cache.GetAll()

	cache.Add("3", String("hahah"))
	cache.GetAll()

	if _, ok := cache.Get("2"); ok {
		cache.GetAll()
	}
	cache.Add("4", String("caonima"))
	cache.GetAll()

	cache.Add("1", String("2345"))
	cache.GetAll()

	if _, ok := cache.Get("1"); ok {
		cache.GetAll()
	}

	cache.RemoveOldest()
	cache.GetAll()

	cache.RemoveOldest()
	cache.GetAll()

	cache.Add("2", String("fuck fuck"))
	cache.GetAll()

	cache.RemoveOldest()
	cache.GetAll()
}
