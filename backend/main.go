package main

import (
	"FlightAPI/crawlers"
	"FlightAPI/handlers"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"time"
)

func main() {
	ctx := context.Background()
	// Initialize Redis client
	// This should be moved to a config file or env var in production code
	rdb := redis.NewClient(&redis.Options{Addr: "redis:6379", Password: "", DB: 0})

	// Create timeout context with 5 min timeout.
	// This context will be used to cancel the API call if it takes too long.
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Create ticker to trigger API calls every 30 minutes
	// This shouldn't be hardcoded in production code. We should pull this from env vars or config files.
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	go func() {
		// Trigger the first API call immediately
		err := crawlers.CallMockyAPI(timeoutCtx, rdb)
		if err != nil {
			log.Printf("Error calling Mocky API on start: %v", err)
		}

		for {
			select {
			case <-ticker.C:
				err := crawlers.CallMockyAPI(timeoutCtx, rdb)
				if err != nil {
					log.Printf("Error calling Mocky API: %v", err)
				}
			case <-timeoutCtx.Done():
				log.Println("Context canceled, stopping ticker")
				return
			}
		}
	}()

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello World!",
		})
	})

	r.POST("/login", LoginHandler)

	// Test JWT authentication
	secret := r.Group("/secret")
	secret.Use(JWTAuthMiddleware())
	secret.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "This is a secret message!",
		})
	})

	protected := r.Group("/api")
	// Middleware to check JWT token
	protected.Use(JWTAuthMiddleware())

	// Middleware to add redis client in to route group
	protected.Use(func(c *gin.Context) {
		c.Set("redisClient", rdb)
		c.Next()
	})

	// Route to fetch all flights from Redis
	protected.GET("/flights", handlers.GetAll)

	// Route to fetch all the dates where flights are available
	protected.GET("/dates", handlers.GetDates)

	// Route to fetch all flights from a date
	protected.GET("/flights/:date", handlers.GetFlightsFromDate)

	err := r.Run(":8080")
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
