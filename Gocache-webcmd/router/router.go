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
	r.POST("/connect", h.Connect())
	r.POST("/disconnect", h.CheckConnect(), h.Disconnect())
	r.POST("/search", h.CheckConnect(), h.Search())
	server := r.Group("/server")
	{
		server.GET("/", h.GetServerInfoHTML())
		server.GET("/refresh", h.CheckConnect(), h.UsedBytesMiddle(), h.Refresh())
		server.GET("/info", h.CheckConnect(), h.UsedBytesMiddle(), h.Info())
	}

	cache := r.Group("/caches")
	{
		cache.GET("/", h.GetCachesHTML())
		cache.GET("/list", h.CheckConnect(), h.CheckPeer(), h.GetCachesList())
		cache.POST("/delete", h.CheckConnect(), h.DeleteCache())
		cache.POST("/update", h.CheckConnect(), h.UpdateCache())
		cache.POST("/set", h.CheckConnect(), h.SetCache())
		cache.POST("/get", h.CheckConnect(), h.GetCache())
	}
	peer := r.Group("/peers")
	{
		peer.GET("/", h.GetPeerHTML())
		peer.GET("/list", h.CheckConnect(), h.GetPeerList())
		peer.POST("/connect", h.CheckConnect(), h.ConnectNewPeer())
		peer.POST("/delete", h.CheckConnect(), h.DeletePeer())
	}
	return r
}
