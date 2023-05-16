package main

import (
	"fmt"
	"log"

	"github.com/meguriri/GoCache/client/client"
)

func main() {
	c := client.Client{}
	server, err := c.Connect("127.0.0.1:8080")
	if err != nil {
		log.Println("dial err", err.Error())
		return
	}
	fmt.Println(*server)
	c.CMD()
}
