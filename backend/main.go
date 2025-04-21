package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello World!",
		})
	})

	r.POST("/login", LoginHandler)

	secret := r.Group("/secret")
	secret.Use(JWTAuthMiddleware())
	secret.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "This is a secret message!",
		})
	})

	r.Run(":8080")
}
