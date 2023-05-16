package router

import (
	"github.com/gin-gonic/gin"
	h "github.com/meguriri/Gocache/webcmd/handler"
)

func InitRouter() *gin.Engine {
	r := gin.Default()
	r.Static("/static", "./static/")
	r.LoadHTMLGlob("templates/*")

	r.GET("/", h.GetConnectHTML())
	connect := r.Group("/connect")
	{
		conn
	}
	cache := r.Group("/caches")
	{
		cache.GET("/", h.GetCacheHTML())
	}
	peer := r.Group("/peers")
	{
		peer.GET("/", h.GetPeerHTML())
	}
	setting := r.Group("/setting")
	{
		setting.GET("/", h.GetSettingHTML())
	}
	return r
}
