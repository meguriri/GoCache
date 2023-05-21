package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meguriri/Gocache/webcmd/data"
)

func GetConnectHTML() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	}
}

func Refresh() gin.HandlerFunc {
	return func(c *gin.Context) {
		usedBytes, _ := c.Get("usedBytes")
		totalBytes, _ := c.Get("cacheBytes")
		peersNumber, _ := c.Get("peersNumber")
		c.JSON(http.StatusOK, gin.H{
			"peersNumber": peersNumber,
			"usedBytes":   usedBytes,
			"totalBytes":  totalBytes,
		})
	}
}

func Info() gin.HandlerFunc {
	return func(c *gin.Context) {
		usedBytes, _ := c.Get("usedBytes")
		totalBytes, _ := c.Get("cacheBytes")
		res, err := data.Client.GetServerInfo()
		if err != nil {
			log.Println("[connectHandler.Info] err", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":  500,
				"error": err.Error(),
			})
			return
		}
		ma := make(map[string]interface{})
		json.Unmarshal(res, &ma)
		log.Println("ma", ma)
		if len(ma) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":  500,
				"error": "info is nil",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":       200,
			"res":        "connect server success",
			"ip":         ma["ip"].(string),
			"port":       ma["port"].(string),
			"peers":      int(ma["peers"].(float64)),
			"policy":     ma["policy"].(string),
			"usedBytes":  usedBytes,
			"totalBytes": totalBytes,
		})
	}
}

func Connect() gin.HandlerFunc {
	return func(c *gin.Context) {
		var address string
		address, err := c.Cookie("address")
		if err != nil {
			log.Println("[connectHandler.connect] cookie err", err)
			address = c.PostForm("address")
		}
		log.Println("[connectHandler Connect] address", address)
		data.Client.Address = address
		res, err := data.Client.Connect()
		if err != nil {
			c.SetCookie("address", "", -1, "/", "", false, false)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"res":  "connect server error" + err.Error(),
			})
			return
		}
		c.SetCookie("address", address, 3600, "/", "", false, false)
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"res":  res,
		})
	}
}

func Disconnect() gin.HandlerFunc {
	return func(c *gin.Context) {
		res, err := data.Client.Exit()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"res":  "disconnect server error" + err.Error(),
			})
			return
		}
		c.SetCookie("address", "abc", -1, "/", "", false, false)
		address, err := c.Cookie("address")
		if err != nil {
			log.Println("[Disconnect] cookie err", err)
		} else {
			log.Println("[Disconnect] cookie ", address)
		}
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"res":  res,
		})
	}
}

func Search() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"search": "fuck",
		})
	}
}
