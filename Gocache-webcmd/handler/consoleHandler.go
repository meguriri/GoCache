package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetConsoleHTML() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "console.html", nil)
	}
}
