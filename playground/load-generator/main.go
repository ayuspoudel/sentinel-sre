package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

func main() {
	targetURL := os.Getenv("TARGET_URL")
	if targetURL == "" {
		log.Fatal("TARGET_URL environment variable is not set")
	}

	concurrency := getEnvInt("CONCURRENCY", 10)
	sleepMs := getEnvInt("SLEEP_MS", 50)

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	log.Println("load generator started")
	log.Println("target:", targetURL)
	log.Println("concurrency:", concurrency)
	log.Println("sleep(ms):", sleepMs)

	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func(workerID int) {
			defer wg.Done()

			for {
				resp, err := client.Get(targetURL)
				if err != nil {
					log.Printf("worker %d error: %v\n", workerID, err)
				} else {
					resp.Body.Close()
				}

				time.Sleep(time.Duration(sleepMs) * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
}

func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}

	parsed, err := strconv.Atoi(val)
	if err != nil {
		log.Fatalf("invalid value for %s: %s", key, val)
	}

	return parsed
}
