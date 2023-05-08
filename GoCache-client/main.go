package main

import (
	"log"

	"github.com/meguriri/GoCache/client/tcp"
)

func main() {
	conn := tcp.Dial("127.0.0.1:8080")
	if conn == nil {
		log.Println("dial err")
		return
	}
	tcp.CMD(conn)
}
