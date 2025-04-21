package handlers

import (
	"FlightAPI/models"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func GetAllFromRedis(c *gin.Context) {
	log.Println("Starting GetAllFromRedis")

	// Retrieve the Redis client from the Gin context
	rdb, ok := c.MustGet("redisClient").(*redis.Client)
	if !ok {
		log.Println("Redis client not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Redis client not found"})
		return
	}

	// Use SCAN to fetch keys
	var cursor uint64
	var keys []string
	for {
		scanKeys, newCursor, err := rdb.Scan(ctx, cursor, "*", 10).Result()
		if err != nil {
			log.Printf("Error scanning keys: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan keys from Redis"})
			return
		}
		keys = append(keys, scanKeys...)
		cursor = newCursor
		if cursor == 0 {
			break
		}
	}

	log.Printf("Keys retrieved: %v", keys)

	if len(keys) == 0 {
		log.Println("No keys found in Redis")
		c.JSON(http.StatusOK, gin.H{"message": "No keys found in Redis"})
		return
	}

	// Fetch and map flight data concurrently
	var flights []models.Flight
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, key := range keys {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()

			keyType, err := rdb.Type(ctx, key).Result()
			if err != nil {
				log.Printf("Failed to fetch type for key %s: %v", key, err)
				return
			}

			log.Printf("Key: %s, Type: %s", key, keyType)

			if keyType == "list" {
				// Fetch list data
				listData, err := rdb.LRange(ctx, key, 0, -1).Result()
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
					mu.Lock()
					flights = append(flights, flight)
					mu.Unlock()
				}
			} else if keyType == "hash" {
				// Handle hash data
				hashData, err := rdb.HGetAll(ctx, key).Result()
				if err != nil {
					log.Printf("Failed to fetch hash for key %s: %v", key, err)
					return
				}

				log.Printf("Hash data for key %s: %v", key, hashData)

				if len(hashData) == 0 {
					log.Printf("Hash data for key %s is empty", key)
					return
				}

				var flight models.Flight
				if err := mapToStruct(hashData, &flight); err != nil {
					log.Printf("Failed to map hash data to struct for key %s: %v", key, err)
					return
				}

				mu.Lock()
				flights = append(flights, flight)
				mu.Unlock()
			} else {
				log.Printf("Skipping unsupported key type for key %s: %s", key, keyType)
			}
		}(key)
	}

	wg.Wait()

	log.Printf("Returning flights")
	c.JSON(http.StatusOK, flights)
}

// Helper function to map Redis hash data to a struct
func mapToStruct(hashData map[string]string, dest interface{}) error {
	// Convert the hash data to JSON
	jsonData, err := json.Marshal(hashData)
	if err != nil {
		return err
	}
	// Unmarshal the JSON into the destination struct
	return json.Unmarshal(jsonData, dest)
}
