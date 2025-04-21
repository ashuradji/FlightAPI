package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http"
)

func main() {

	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{Addr: "redis:6379", Password: "", DB: 0})

	go callMockyAPI(ctx, rdb)

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
