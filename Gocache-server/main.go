package main

import (
	"log"

	"github.com/meguriri/GoCache/server/manager"
)

func main() {
	if err := manager.StartServer(); err != nil {
		log.Fatalf(err.Error())
	}
}
