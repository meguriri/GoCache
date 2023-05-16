package client

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func (c *Client) CMD() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(c.Conn.RemoteAddr().String(), ">")
		b, _, _ := reader.ReadLine()
		r := strings.Trim(string(b), " ")
		req := strings.Split(r, " ")
		if (req[0] == "set" || req[0] == "SET") && len(req) == 3 {
			res, err := c.Set(req[1], req[2])
			if err != nil {
				fmt.Println("set error", err)
				continue
			}
			fmt.Println(res)
		} else if (req[0] == "get" || req[0] == "GET") && len(req) == 2 {
			res, err := c.Get(req[1])
			if err != nil {
				fmt.Println("get error", err)
				continue
			}
			fmt.Println(res)
		} else if (req[0] == "del" || req[0] == "DEL") && len(req) == 2 {
			res, err := c.Del(req[1])
			if err != nil {
				fmt.Println("del error", err)
				continue
			}
			fmt.Println(res)
		} else if req[0] == "exit" || req[0] == "EXIT" && len(req) == 1 {
			res, err := c.Exit()
			if err != nil {
				fmt.Println("exit error", err)
				continue
			}
			fmt.Println(res)
			return
		} else if (req[0] == "kill" || req[0] == "KILL") && len(req) == 2 {
			res, err := c.Kill(req[1])
			if err != nil {
				fmt.Println("kill error", err)
				continue
			}
			fmt.Println(res)
		} else if (req[0] == "connect" || req[0] == "CONNECT") && len(req) == 4 {
			res, err := c.ConnectPeer(req[1], req[2], req[3])
			if err != nil {
				fmt.Println("connect error", err)
				continue
			}
			fmt.Println(res)
		} else if (req[0] == "getall" || req[0] == "GETALL") && len(req) == 2 {
			res, err := c.GetAllCache(req[1])
			if err != nil {
				fmt.Println("getall error", err)
				continue
			}
			for k, v := range res {
				fmt.Printf("%s:%s\n", k, v)
			}
		} else if (req[0] == "info" || req[0] == "INFO") && len(req) == 2 {
			res, err := c.Info(req[1])
			if err != nil {
				fmt.Println("info error", err)
				continue
			}
			fmt.Printf("name: %s\n", res["name"].(string))
			fmt.Printf("address: %s\n", res["address"].(string))
			fmt.Printf("replacement: %s\n", res["replacement"].(string))
			fmt.Printf("cacheBytes: %d\n", int(res["cacheBytes"].(float64)))
			fmt.Printf("usedBytes: %d\n", int(res["usedBytes"].(float64)))
		} else if (req[0] == "peers" || req[0] == "PEERS") && len(req) == 1 {
			res, err := c.GetAllPeers()
			if err != nil {
				fmt.Println("peers error", err)
				continue
			}
			for i, v := range res {
				fmt.Printf("%d: %s\n", i, v)
			}
		} else {
			fmt.Println("error input,Please input the correct oreder!")
		}
	}
}
