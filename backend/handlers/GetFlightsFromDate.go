package handlers

import (
	"FlightAPI/models"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"sort"
	"sync"
)

func GetFlightsFromDate(ctx *gin.Context) {
	// Extract date from URL parameter
	date := ctx.Param("date")
	log.Printf("Received date parameter: %s", date)

	// Retrieve Redis client from context
	redisClient, exists := ctx.Get("redisClient")
	if !exists {
		log.Println("Redis client not found in context")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Redis client not found"})
		return
	}

	rdb, ok := redisClient.(*redis.Client)
	if !ok {
		log.Println("Invalid Redis client type")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid Redis client"})
		return
	}

	// Check the type of the key in Redis
	keyType, err := rdb.Type(context.Background(), date).Result()
	if err != nil {
		log.Printf("Error fetching key type from Redis for key '%s': %v", date, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch key type from Redis"})
		return
	}

	var flights []models.Flight
	flightChan := make(chan models.Flight)
	var wg sync.WaitGroup

	switch keyType {
	case "list":
		// Handle list type
		data, err := rdb.LRange(context.Background(), date, 0, -1).Result()
		if err != nil {
			log.Printf("Error fetching list data from Redis for key '%s': %v", date, err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch flights from Redis"})
			return
		}
		for _, item := range data {
			wg.Add(1)
			go func(item string) {
				defer wg.Done()
				var flight models.Flight
				if err := json.Unmarshal([]byte(item), &flight); err != nil {
					log.Printf("Error parsing flight data: %v", err)
					return
				}
				flightChan <- flight
			}(item)
		}

	case "hash":
		// Handle hash type
		data, err := rdb.HGetAll(context.Background(), date).Result()
		if err != nil {
			log.Printf("Error fetching hash data from Redis for key '%s': %v", date, err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch flights from Redis"})
			return
		}
		for _, item := range data {
			wg.Add(1)
			go func(item string) {
				defer wg.Done()
				var flight models.Flight
				if err := json.Unmarshal([]byte(item), &flight); err != nil {
					log.Printf("Error parsing flight data: %v", err)
					return
				}
				flightChan <- flight
			}(item)
		}

	default:
		log.Printf("Unsupported Redis key type '%s' for key '%s'", keyType, date)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported Redis key type"})
		return
	}

	// Close the channel once all goroutines are done
	go func() {
		wg.Wait()
		close(flightChan)
	}()

	// Collect flights from the channel
	for flight := range flightChan {
		flights = append(flights, flight)
	}

	// Sort flights by departure time (earliest to latest)
	sort.Slice(flights, func(i, j int) bool {
		return flights[i].DepartureTime < flights[j].DepartureTime
	})

	// Return sorted flights as JSON
	ctx.JSON(http.StatusOK, gin.H{"flights": flights})
}
