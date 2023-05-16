package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/meguriri/GoCache/client/server"
)

type Client struct {
	Conn net.Conn
	Res  []byte
}

func (c *Client) Connect(addr string) (*server.Server, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	log.Println("connect success", conn)
	c.Conn = conn
	c.Res = make([]byte, 1024*1024)
	n, _ := conn.Read(c.Res)
	ma := make(map[string]interface{})
	json.Unmarshal(c.Res[:n], &ma)
	return &server.Server{Ip: ma["ip"].(string), Port: ma["port"].(string), Peers: int(ma["peers"].(float64)), Policy: ma["policy"].(string)}, nil
}

func (c *Client) Exit() (string, error) {
	defer c.Conn.Close()
	_, err1 := c.Conn.Write([]byte("exit"))
	n, err2 := c.Conn.Read(c.Res)
	if err1 != nil {
		return "write error", err1
	}
	if err2 != nil {
		return "read error", err2
	}
	return string(c.Res[:n]), nil
}

func (c *Client) Set(key, value string) (string, error) {
	_, err1 := c.Conn.Write([]byte("set " + key + " " + value))
	n, err2 := c.Conn.Read(c.Res)
	if err1 != nil {
		return "write error", err1
	}
	if err2 != nil {
		return "read error", err2
	}
	return string(c.Res[:n]), nil
}

func (c *Client) Get(key string) (string, error) {
	_, err1 := c.Conn.Write([]byte("get " + key))
	n, err2 := c.Conn.Read(c.Res)
	if err1 != nil {
		return "write error", err1
	}
	if err2 != nil {
		return "read error", err2
	}
	return string(c.Res[:n]), nil
}

func (c *Client) Del(key string) (string, error) {
	_, err1 := c.Conn.Write([]byte("del " + key))
	n, err2 := c.Conn.Read(c.Res)
	if err1 != nil {
		return "write error", err1
	}
	if err2 != nil {
		return "read error", err2
	}
	return string(c.Res[:n]), nil
}

func (c *Client) Kill(address string) (string, error) {
	_, err1 := c.Conn.Write([]byte("kill " + address))
	n, err2 := c.Conn.Read(c.Res)
	if err1 != nil {
		return "write error", err1
	}
	if err2 != nil {
		return "read error", err2
	}
	return string(c.Res[:n]), nil
}

func (c *Client) ConnectPeer(name, address, cacheBytes string) (string, error) {
	_, err1 := c.Conn.Write([]byte("connect " + name + " " + address + " " + cacheBytes))
	n, err2 := c.Conn.Read(c.Res)
	if err1 != nil {
		return "write error", err1
	}
	if err2 != nil {
		return "read error", err2
	}
	return string(c.Res[:n]), nil
}

func (c *Client) GetAllCache(address string) (map[string]string, error) {
	_, err1 := c.Conn.Write([]byte("getall " + address))
	n, err2 := c.Conn.Read(c.Res)
	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		return nil, err2
	}
	if string(c.Res[:n]) == "null" {
		return nil, fmt.Errorf("%s is not exist", address)
	}
	l := make(map[string]string)
	s := strings.Trim(string(c.Res[:n]), "\000")
	json.Unmarshal([]byte(s), &l)
	log.Println(s)
	return l, nil
}

func (c *Client) Info(address string) (map[string]interface{}, error) {
	_, err1 := c.Conn.Write([]byte("info " + address))
	n, err2 := c.Conn.Read(c.Res)
	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		return nil, err2
	}
	if string(c.Res[:n]) == "null" {
		return nil, fmt.Errorf("%s is not exist", address)
	}
	l := make(map[string]interface{})
	s := strings.Trim(string(c.Res[:n]), "\000")
	json.Unmarshal([]byte(s), &l)
	return l, nil
}

func (c *Client) GetAllPeers() ([]string, error) {
	_, err1 := c.Conn.Write([]byte("peers"))
	n, err2 := c.Conn.Read(c.Res)
	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		return nil, err2
	}
	l := make([]string, 0)
	s := strings.Trim(string(c.Res[:n]), "\000")
	json.Unmarshal([]byte(s), &l)
	return l, nil
}
