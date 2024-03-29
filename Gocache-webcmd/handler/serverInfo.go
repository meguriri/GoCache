package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetServerInfoHTML() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "serverInfo.html", nil)
	}
}
