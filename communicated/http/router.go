package http

import "github.com/gin-gonic/gin"

func RouterInit() *gin.Engine {
	r := gin.Default()
	cache := r.Group("/cache")
	{
		cache.GET("/:group/:key", GetKey())
	}
	return r
}
