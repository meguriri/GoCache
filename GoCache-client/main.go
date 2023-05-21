package main

import (
	"fmt"
	"log"

	"github.com/meguriri/GoCache/client/client"
)

func main() {
	c := client.Client{}
	c.Address = "127.0.0.1:8080"
	res, err := c.Connect()
	if err != nil {
		log.Println("dial err", err.Error())
		return
	}
	fmt.Println(res)
	c.CMD()
}
