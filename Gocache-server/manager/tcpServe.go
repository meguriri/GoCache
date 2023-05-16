package manager

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/meguriri/GoCache/server/replacement"
)

func (m *Manager) TCPServe(ctx context.Context, wg *sync.WaitGroup) {
	listen, err := net.Listen("tcp", m.addr)
	if err != nil {
		log.Printf("listen err=%v\n", err)
		return
	}
	log.Println("listen on ", m.addr)
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Printf("Accept() err=%v\n", err)
		} else {
			log.Printf("client ip=%v\n", conn.RemoteAddr().String())
		}
		ma := make(map[string]interface{})
		ma["ip"] = ManagerIP
		ma["port"] = ManagerPort
		ma["peers"] = len(m.cachePeers)
		ma["policy"] = replacement.ReplacementPolicy
		b, _ := json.Marshal(ma)
		conn.Write(b)
		go m.TCPhandler(ctx, conn)
	}
}

func (m *Manager) TCPhandler(ctx context.Context, conn net.Conn) {
	for {
		buf := make([]byte, 1024*1024)
		_, err := conn.Read(buf)
		req := strings.Trim(string(buf), "\000")
		if err != nil {
			log.Println(conn.RemoteAddr().String(), "is leave")
			return
		}
		li := strings.Split(req, " ")
		var resp = []byte{}
		if li[0] == "set" {
			if ok := m.Set(ctx, li[1], []byte(li[2])); !ok {
				resp = []byte("set error")
			} else {
				resp = []byte("OK")
			}
		} else if li[0] == "get" {
			res, err := m.Get(ctx, li[1])
			if err != nil {
				resp = []byte("(nil)")
			} else {
				resp = []byte("\"" + string(res) + "\"")
			}
		} else if li[0] == "del" {
			res := m.Del(ctx, li[1])
			if !res {
				resp = []byte("(integer) 0")
			} else {
				resp = []byte("(integer) 1")
			}
		} else if li[0] == "exit" {
			resp = []byte("bye!")
			conn.Write(resp)
			log.Println(conn.RemoteAddr().String(), "is leave")
			conn.Close()
			break
		} else if li[0] == "kill" {
			if ok, err := m.Kill(ctx, li[1]); ok {
				resp = []byte(li[1] + " is logout")
			} else {
				resp = []byte(li[1] + "logout err:" + err.Error())
			}
		} else if li[0] == "connect" {
			bytes, _ := strconv.Atoi(li[3])
			if ok := m.Connect(li[1], li[2], int64(bytes)); ok {
				resp = []byte(li[2] + " is connected")
			} else {
				resp = []byte(li[2] + " connect err:")
			}
		} else if li[0] == "getall" {
			res := m.GetAllCache(ctx, li[1])
			for _, v := range res {
				log.Println("res:", string(v))
			}
			r, _ := json.Marshal(res)
			resp = r
		} else if li[0] == "info" {
			res := m.Info(ctx, li[1])
			r, _ := json.Marshal(res)
			log.Println(res)
			resp = r
		} else if li[0] == "peers" {
			res := m.GetAllPeerAddress()
			r, _ := json.Marshal(res)
			log.Println(res)
			resp = r
		} else {
			resp = []byte("error input")
		}
		conn.Write(resp)
	}
}
