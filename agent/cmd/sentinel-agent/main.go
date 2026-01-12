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
	"github.com/ayuspoudel/sentinel-sre/agent/internal/heartbeat"
	"github.com/ayuspoudel/sentinel-sre/agent/internal/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Agent controller config (heartbeat target)
	agentControllerURL := os.Getenv("AGENT_CONTROLLER_URL")
	clusterName := os.Getenv("SENTINEL_CLUSTER_NAME")
	agentVersion := os.Getenv("SENTINEL_AGENT_VERSION")

	if agentControllerURL == "" {
		log.Fatal("AGENT_CONTROLLER_URL must be set")
	}
	if clusterName == "" {
		log.Fatal("SENTINEL_CLUSTER_NAME must be set")
	}
	if agentVersion == "" {
		log.Fatal("SENTINEL_AGENT_VERSION must be set")
	}
	agentID := os.Getenv("SENTINEL_AGENT_ID")
	if agentID == "" {
		log.Fatal("SENTINEL_AGENT_ID must be set")
	}
	// Sentinel control plane client (admission only)
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

	// Agent → Agent Controller heartbeat (independent)
	go heartbeat.Start(
		ctx,
		30*time.Second,
		func(ctx context.Context) error {
			return heartbeat.Send(
				ctx,
				agentControllerURL,
				agentID,
				clusterName,
				agentVersion,
			)
		},
	)

	<-stop
	log.Println("sentinel-agent shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
}
