package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetSettingHTML() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "setting.html", nil)
	}
}
