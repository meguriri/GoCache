package main

import "github.com/meguriri/Gocache/webcmd/router"

func main() {
	r := router.InitRouter()
	r.Run(":8888")
}
