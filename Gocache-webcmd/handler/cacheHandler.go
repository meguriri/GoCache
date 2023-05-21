package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meguriri/Gocache/webcmd/data"
)

type cache struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func GetCachesHTML() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "cachelist.html", nil)
	}
}

func GetCachesList() gin.HandlerFunc {
	return func(c *gin.Context) {
		address := c.GetString("address")
		res, err := data.Client.GetAllCache(address)
		if err != nil {
			log.Println("GetAllCache err:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":  500,
				"error": err.Error(),
			})
			return
		}
		log.Println("[cacheHandler.GetCachesList] res", res)
		list := make([]cache, 0)
		for k, v := range res {
			list = append(list, cache{Key: k, Value: v})
		}
		log.Println("[cacheHandler.GetCachesList] list", list)
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"list": list,
		})
	}
}

func DeleteCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.PostForm("key")
		res, err := data.Client.Del(key)
		log.Println("[cacheHandler.DeleteCache] res", res)
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

func UpdateCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		key, value := c.PostForm("key"), c.PostForm("value")
		res, err := data.Client.Set(key, value)
		log.Println("[cacheHandler.UpdateCache] res", res)
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

func SetCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		key, value := c.PostForm("key"), c.PostForm("value")
		res, err := data.Client.Set(key, value)
		log.Println("[cacheHandler.SetCache] res", res)
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

func GetCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.PostForm("key")
		res, err := data.Client.Get(key)
		log.Println("[cacheHandler.GetCache] res", res)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":  500,
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":  200,
			"key":   key,
			"value": res,
		})
	}
}
