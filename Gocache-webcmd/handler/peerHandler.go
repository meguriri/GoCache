package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetPeerHTML() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "peerlist.html", nil)
	}
}
