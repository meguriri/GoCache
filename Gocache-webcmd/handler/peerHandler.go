package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meguriri/Gocache/webcmd/data"
)

func GetPeerHTML() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "peerlist.html", nil)
	}
}

func GetPeerList() gin.HandlerFunc {
	return func(c *gin.Context) {
		peerlist, err := data.Client.GetAllPeers()
		log.Println("[peerHandler.GetPeerList] peerlist", peerlist)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":  500,
				"error": err.Error(),
			})
			return
		}

		list := make([]map[string]interface{}, 0)
		for _, address := range peerlist {
			info, err := data.Client.Info(address)
			if err != nil {
				log.Println(address, "err", err.Error())
				continue
			}
			list = append(list, info)
		}
		log.Println("[peerHandler.GetPeerList]", list)
		c.JSON(http.StatusOK, gin.H{
			"code":  200,
			"error": nil,
			"list":  list,
		})
	}
}

func ConnectNewPeer() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.PostForm("name")
		address := c.PostForm("address")
		maxBytes := c.PostForm("maxBytes")
		res, err := data.Client.ConnectPeer(name, address, maxBytes)
		log.Println("[peerHandler.ConnectNewPeer] connect res", res)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":  500,
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"res":  res,
		})
	}
}

func DeletePeer() gin.HandlerFunc {
	return func(c *gin.Context) {
		address := c.PostForm("address")
		res, err := data.Client.Kill(address)
		log.Println("[peerHandler.DeletePeer]", res)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":  500,
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"res":  res,
		})
	}
}
