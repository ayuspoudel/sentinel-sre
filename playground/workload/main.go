package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total Number of HTTP Requests",
		},
		[]string{"status"},
	)
)

func main() {
	rand.Seed(time.Now().UnixNano())
	prometheus.MustRegister(httpRequestsTotal)

	// simulate a fake business endpoint

	http.HandleFunc("/work", func(w http.ResponseWriter, _ *http.Request) {
		if rand.Float64() < 0.02 {
			httpRequestsTotal.WithLabelValues("500").Inc()
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error"))
			return
		}
		httpRequestsTotal.WithLabelValues("200").Inc()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	log.Println("workload service listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
