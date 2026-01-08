package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent"
	"github.com/ayuspoudel/sentinel-sre/controlplane/registryClient"
)

func main() {
	registryURL := os.Getenv("REGISTRY_URL")
	if registryURL == "" {
		log.Fatal("REGISTRY_URL is required")
	}

	reconcileInterval := 10 * time.Second
	if v := os.Getenv("RECONCILE_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			reconcileInterval = d
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	registryClient := registryClient.New(registryURL)

	controller := agent.NewController(registryClient, reconcileInterval)

	go controller.Run(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	log.Println("shutdown signal received")

	cancel()

	time.Sleep(1 * time.Second)
	log.Println("sentinel agent controller exited cleanly")
}
