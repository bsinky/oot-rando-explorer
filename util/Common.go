package util

import (
	"net/http"

	"github.com/bsinky/sohrando/authentication"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ViewModel struct {
	User *authentication.UserDisplay
}

func ViewData(c *gin.Context, data *gin.H) *gin.H {
	(*data)["User"] = authentication.GetCurrentUser(c)
	return data
}

func ConnectDatabase(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("database", db)
	}
}

func GetDatabase(c *gin.Context) *gorm.DB {
	return c.Value("database").(*gorm.DB)
}

func HtmxRedirect(c *gin.Context, dst string) {
	c.Status(http.StatusOK)
	c.Header("HX-Location", dst)
}
