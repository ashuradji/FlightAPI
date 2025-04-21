package main

import (
	"FlightAPI/models"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"time"
)

func callMockyAPI(ctx context.Context, rdb *redis.Client) {
	// Mocky API URL: This shouldn't be hardcoded in production code
	url := "https://run.mocky.io/v3/60991ebd-1a38-4b8c-9e29-6466adb66fc6"

	// Create a new HTTP client
	client := &http.Client{}

	// Create a new GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Set the request header
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Read response body
	decoder := json.NewDecoder(resp.Body)

	_, err = decoder.Token()

	if err != nil {
		log.Fatal("Failed to read start object %v", err)
	}

	for decoder.More() {
		tok, err := decoder.Token()
		if err != nil {
			log.Fatalf("Failed to read key: %v", err)
		}

		if key, ok := tok.(string); ok && key == "flights" {
			// Step 3: Read `[`
			_, err := decoder.Token()
			if err != nil {
				log.Fatalf("Failed to read flights array start: %v", err)
			}

			// Step 4: Stream array items
			for decoder.More() {
				var flight models.Flight
				if err := decoder.Decode(&flight); err != nil {
					log.Printf("Decode error: %v", err)
					continue
				}
				// Save flight to Redis
				err := saveFlightByDate(ctx, rdb, flight)

				if err != nil {
					log.Printf("Error saving flight: %v", err)
					continue
				}
			}

			// Step 5: Read `]`
			_, _ = decoder.Token()
			break
		}
	}
}

func saveFlightByDate(ctx context.Context, rdb *redis.Client, flight models.Flight) error {
	// Parse and format the date
	t, err := time.Parse(time.RFC3339, flight.DepartureTime)
	if err != nil {
		return fmt.Errorf("invalid departure time: %w", err)
	}
	dateKey := t.Format("2006-01-02")
	redisKey := fmt.Sprintf("flights:%s", dateKey)

	// Convert to JSON
	data, err := json.Marshal(flight)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	// LPUSH for most recent first
	return rdb.LPush(ctx, redisKey, data).Err()
}
