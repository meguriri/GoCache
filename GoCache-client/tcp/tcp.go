package tcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

func CMD(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(conn.RemoteAddr().String(), ">")
		req, _, _ := reader.ReadLine()
		conn.Write([]byte(req))
		res := make([]byte, 1024*1024)
		_, err := conn.Read(res)
		if err != nil {
			fmt.Println(conn.RemoteAddr().String(), "is disconnected")
			return
		}
		if string(req) == "exit" {
			fmt.Println(string(res))
			break
		} else if strings.HasPrefix(string(req), "getall") {
			l := make([][]byte, 0)
			s := strings.Trim(string(res), "\000")
			err := json.Unmarshal([]byte(s), &l)
			if err != nil {
				fmt.Println(err)
				continue
			}
			for _, v := range l {
				fmt.Printf("%s", string(v))
			}
		} else if strings.HasPrefix(string(req), "info") {
			l := make(map[string]interface{})
			s := strings.Trim(string(res), "\000")
			err := json.Unmarshal([]byte(s), &l)
			if err != nil {
				fmt.Println(err)
				continue
			}
			for k, v := range l {
				fmt.Println(k, ":", v)
			}
		} else if strings.HasPrefix(string(req), "peers") {
			l := make([]string, 0)
			s := strings.Trim(string(res), "\000")
			err := json.Unmarshal([]byte(s), &l)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(len(l), "peers:")
			for i, v := range l {
				fmt.Println(i, ":", v)
			}
		} else {
			fmt.Println(string(res))
		}
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
