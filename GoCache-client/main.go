package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func CMD(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("127.0.0.1:8080>")
		req, _, _ := reader.ReadLine()
		conn.Write([]byte(req))
		res := make([]byte, 1024)
		conn.Read(res)
		fmt.Println(string(res))
	}
}

func Dial(address string) net.Conn {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("client dial err=", err)
		return nil
	}
	fmt.Println("connect success", conn)
	return conn
}

func main() {
	conn := Dial("127.0.0.1:8080")
	if conn == nil {
		log.Println("dial err")
		return
	}
	CMD(conn)
}
