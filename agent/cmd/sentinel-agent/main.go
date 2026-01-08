package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ayuspoudel/sentinel-sre/agent/internal/admission"
	"github.com/ayuspoudel/sentinel-sre/agent/internal/client"
	"github.com/ayuspoudel/sentinel-sre/agent/internal/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	sentinel, err := client.NewSentinelClient()
	if err != nil {
		log.Fatalf("startup failed: %v", err)
	}
	if err := sentinel.HealthCheck(ctx); err != nil {
		log.Printf("WARNING: sentinel not reachable at startup: %v", err)
	} else {
		log.Printf("Sentinel control plane reachable")
	}
	admissionHandler := admission.NewHandler(sentinel)
	// Admission server
	srv := server.NewServer(":8443", admissionHandler)

	go func() {
		log.Println("sentinel-agent admission server starting on :8443")
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("agent server error: %v", err)
		}
	}()

	<-stop
	log.Println("sentinel-agent shutting  down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
}
