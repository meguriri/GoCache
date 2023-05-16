package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meguriri/Gocache/webcmd/data"
)

func GetConnectHTML() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	}
}

func Connect() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.PostForm("ip")
		port := c.PostForm("port")
		Type := c.PostForm("type")
		if Type == "connect" {
			conn, err := data.Client.Connect(ip + ":" + port)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code": "500",
					"res":  "connect server error" + err.Error(),
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": "200",
				"res":  "connect server success",
			})
		} else if Type == "disconnect" {

		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": "500",
			"res":  "type error",
		})
	}
}

func Search() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	}
}
