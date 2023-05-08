package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetCacheHTML() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "cachelist.html", nil)
	}
}
