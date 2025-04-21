package handlers

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http"
	"sort"
	"sync"
	"time"
)

func GetDates(ctx *gin.Context) {
	// Retrieve Redis client from context
	redisClient, exists := ctx.Get("redisClient")
	if !exists {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Redis client not found"})
		return
	}

	rdb, ok := redisClient.(*redis.Client)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid Redis client"})
		return
	}

	// Fetch all keys from Redis
	keys, err := rdb.Keys(context.Background(), "*").Result()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch keys from Redis"})
		return
	}

	// Channel to collect parsed dates
	dateChan := make(chan time.Time, len(keys))
	var wg sync.WaitGroup

	// Parse keys concurrently
	for _, key := range keys {
		wg.Add(1)
		go func(k string) {
			defer wg.Done()
			parsedDate, err := time.Parse("2006-01-02", k)
			if err == nil {
				dateChan <- parsedDate
			}
		}(key)
	}

	// Wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(dateChan)
	}()

	// Collect dates from the channel
	var dates []time.Time
	for date := range dateChan {
		dates = append(dates, date)
	}

	// Sort dates from closest to farthest
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	// Convert dates back to string format
	var sortedDates []string
	for _, date := range dates {
		sortedDates = append(sortedDates, date.Format("2006-01-02"))
	}

	// Return sorted dates as JSON
	ctx.JSON(http.StatusOK, gin.H{"dates": sortedDates})
}
