package handlers

import (
	"FlightAPI/models"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func GetFlightsBySearch(ctx *gin.Context) {
	origin := ctx.Query("origin")
	destination := ctx.Query("destination")
	date := ctx.Query("date")

	log.Printf("Received search parameters: origin=%s, destination=%s, date=%s", origin, destination, date)

	// Retrieve the Redis client from the Gin context
	rdb, ok := ctx.MustGet("redisClient").(*redis.Client)
	if !ok {
		log.Println("Redis client not found in context")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Redis client not found"})
		return
	}

	var cursor uint64
	var keys []string

	// Use SCAN to fetch all keys
	for {
		scanKeys, newCursor, err := rdb.Scan(context.Background(), cursor, "*", 10).Result()
		if err != nil {
			log.Printf("Error scanning keys: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan keys from Redis"})
			return
		}
		keys = append(keys, scanKeys...)
		cursor = newCursor
		if cursor == 0 {
			break
		}
	}

	var wg sync.WaitGroup
	flightsChan := make(chan models.Flight, len(keys))

	// Process keys concurrently
	for _, key := range keys {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()

			keyType, err := rdb.Type(context.Background(), key).Result()
			if err != nil {
				log.Printf("Failed to fetch type for key %s: %v", key, err)
				return
			}

			if keyType == "list" {
				// Process list data
				listData, err := rdb.LRange(context.Background(), key, 0, -1).Result()
				if err != nil {
					log.Printf("Failed to fetch list for key %s: %v", key, err)
					return
				}

				for _, item := range listData {
					var flight models.Flight
					if err := json.Unmarshal([]byte(item), &flight); err != nil {
						log.Printf("Failed to unmarshal list item for key %s: %v", key, err)
						continue
					}

					// Filter flights
					if strings.EqualFold(flight.DepartureAirport.Code, origin) &&
						strings.EqualFold(flight.ArrivalAirport.Code, destination) &&
						strings.HasPrefix(flight.DepartureTime, date) {
						flightsChan <- flight
					}
				}
			} else if keyType == "hash" {
				// Process hash data
				hashData, err := rdb.HGetAll(context.Background(), key).Result()
				if err != nil {
					log.Printf("Failed to fetch hash for key %s: %v", key, err)
					return
				}

				if len(hashData) == 0 {
					return
				}

				var flight models.Flight
				if err := mapToStruct(hashData, &flight); err != nil {
					log.Printf("Failed to map hash data to struct for key %s: %v", key, err)
					return
				}

				// Filter flights
				if strings.EqualFold(flight.DepartureAirport.Code, origin) &&
					strings.EqualFold(flight.ArrivalAirport.Code, destination) &&
					strings.HasPrefix(flight.DepartureTime, date) {
					flightsChan <- flight
				}
			}
		}(key)
	}

	// Close the channel once all Goroutines are done
	go func() {
		wg.Wait()
		close(flightsChan)
	}()

	// Collect matching flights from the channel
	var matchingFlights []models.Flight
	for flight := range flightsChan {
		matchingFlights = append(matchingFlights, flight)
	}

	// Return matching flights
	if len(matchingFlights) == 0 {
		ctx.JSON(http.StatusOK, gin.H{"message": "No matching flights found"})
		return
	}

	ctx.JSON(http.StatusOK, matchingFlights)
}
