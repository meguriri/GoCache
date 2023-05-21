package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meguriri/Gocache/webcmd/data"
)

func CheckConnect() gin.HandlerFunc {
	return func(c *gin.Context) {
		if ok, err := data.Client.Heart(); !ok {
			log.Println("[middleware.CheckConnect] Heart err", err)
			_, err := data.Client.Connect()
			if err != nil {
				log.Println("[middleware.CheckConnect] Connect err", err)
				c.Abort()
				return
			}
			c.Next()
		}
		c.Next()
	}
}

func UsedBytesMiddle() gin.HandlerFunc {
	return func(c *gin.Context) {
		usedBytes, totalBytes := 0, 0
		peerList, err := data.Client.GetAllPeers()
		if err != nil {
			log.Println("[middleware.UsedBytesMiddle] UsedBytesMiddle err", err.Error())
			c.Abort()
		}
		for _, address := range peerList {
			m, err := data.Client.Info(address)
			if err != nil {
				log.Println("[middleware.UsedBytesMiddle]", address, "err", err.Error())
				continue
			}
			usedBytes += int(m["usedBytes"].(float64))
			totalBytes += int(m["cacheBytes"].(float64))
		}
		c.Set("peersNumber", len(peerList))
		c.Set("usedBytes", usedBytes)
		c.Set("cacheBytes", totalBytes)
		c.Next()
	}
}

func CheckPeer() gin.HandlerFunc {
	return func(c *gin.Context) {
		address := c.Query("address")
		log.Println("[middleware.CheckPeer] address", address)
		list, err := data.Client.GetAllPeers()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":  500,
				"error": err.Error(),
			})
			c.Abort()
			return
		}
		for _, peer := range list {
			if peer == address {
				log.Println("get address")
				c.Set("address", address)
				c.Next()
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"code":  500,
			"error": "no peer",
		})
		c.Abort()
	}
}
