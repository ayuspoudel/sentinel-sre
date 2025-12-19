package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var sentRequests uint64

func main() {
	targetURL := mustGetEnv("TARGET_URL")
	if targetURL == "" {
		log.Fatal("TARGET_URL environment variable is not set")
	}

	concurrency := getEnvInt("CONCURRENCY", 10)
	sleepMs := getEnvInt("SLEEP_MS", 50)

	transport := &http.Transport{
		MaxIdleConns:        200,
		MaxIdleConnsPerHost: 200,
		IdleConnTimeout:     90 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   2 * time.Second,
	}

	log.Println("load generator started")
	log.Println("target:", targetURL)
	log.Println("concurrency:", concurrency)
	log.Println("sleep(ms):", sleepMs)

	// stats logger
	go logStats()

	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go worker(i, client, targetURL, sleepMs, &wg)
	}

	wg.Wait()
}

func worker(
	id int,
	client *http.Client,
	target string,
	sleepMs int,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for {
		resp, err := client.Get(target)
		if err != nil {
			log.Printf("worker %d error: %v", id, err)
			time.Sleep(200 * time.Millisecond)
			continue
		}

		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		atomic.AddUint64(&sentRequests, 1)

		time.Sleep(time.Duration(sleepMs) * time.Millisecond)
	}
}

func logStats() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		count := atomic.SwapUint64(&sentRequests, 0)
		rps := float64(count) / 5.0

		log.Printf(
			"loadgen stats: requests_last_5s=%d rps=%.2f",
			count,
			rps,
		)
	}
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("%s environment variable is required", key)
	}
	return val
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
