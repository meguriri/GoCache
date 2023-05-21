package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
)

type Client struct {
	Conn    net.Conn
	Address string
	Res     []byte
}

func (c *Client) Connect() (string, error) {
	conn, err := net.Dial("tcp", c.Address)
	if err != nil {
		return "", err
	}
	log.Println("connect success", conn.RemoteAddr())
	c.Conn = conn
	c.Res = make([]byte, 1024*1024)
	n, _ := conn.Read(c.Res)
	return string(c.Res[:n]), nil
}

func (c *Client) GetServerInfo() ([]byte, error) {
	defer c.Conn.Close()
	_, err1 := c.Conn.Write([]byte("server"))
	n, err2 := c.Conn.Read(c.Res)
	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		return nil, err2
	}
	return c.Res[:n], nil
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
	if string(c.Res[:n]) == "connect err" {
		return "connect err", fmt.Errorf("connect err")
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

func (c *Client) Heart() (bool, error) {
	_, err1 := c.Conn.Write([]byte("heart"))
	n, err2 := c.Conn.Read(c.Res)
	if err1 != nil {
		return false, err1
	}
	if err2 != nil {
		return false, err2
	}
	if string(c.Res[:n]) == "OK" {
		return true, nil
	}
	return false, fmt.Errorf("connect is dead")
}
