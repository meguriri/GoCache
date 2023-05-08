package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetConnectHTML() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	}
}
