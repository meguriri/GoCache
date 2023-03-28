package http

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meguriri/GoCache/cache"
)

func GetKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		gName, key := c.Param("group"), c.Param("key")
		fmt.Println(gName, key)
		group := cache.GetGroup(gName)
		if group == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"msg":   "fail",
				"value": nil,
			})
			return
		}
		view, err := group.Get(key)
		log.Println("[GetKey] view ", string(view.GetByte()))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg":   err.Error(),
				"value": nil,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"msg":   "ok",
			"value": string(view.GetByte()),
		})
	}
}
