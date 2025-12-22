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
			Help: "Total number of HTTP requests",
		},
		[]string{"status"},
	)
)

func main() {
	rand.Seed(time.Now().UnixNano())
	prometheus.MustRegister(httpRequestsTotal)

	mux := http.NewServeMux()

	mux.HandleFunc("/work", func(w http.ResponseWriter, _ *http.Request) {
		if rand.Float64() < 0.1 {
			httpRequestsTotal.WithLabelValues("500").Inc()
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error"))
			return
		}

		httpRequestsTotal.WithLabelValues("200").Inc()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: mux,
	}

	log.Println("workload service listening on 0.0.0.0:8080")
	log.Fatal(server.ListenAndServe())
}
