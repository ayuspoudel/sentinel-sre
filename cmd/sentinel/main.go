package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ayuspoudel/sentinel-sre/internal/engine"
	"github.com/ayuspoudel/sentinel-sre/internal/registry"
	"github.com/ayuspoudel/sentinel-sre/internal/server"
)

func main() {
	// Root context controlled by OS signals (Kubernetes-friendly)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Initialize registry (source of truth for guards)
	reg := registry.NewGitRegistry(
		"git@github.com:ayuspoudel/sentinel-manifests.git",
		"develop",
		"/tmp/sentinel",
		registry.GitAuth{
			// SSHKeyPath: "/etc/sentinel/git/id_rsa", // or TokenEnv: "GIT_TOKEN"
		},
	)

	// Initialize Sentinel engine
	eng := engine.New(reg, 10*time.Second, 1*time.Minute)

	// Start engine in background
	go func() {
		if err := eng.Start(ctx); err != nil {
			log.Fatalf("engine error: %v", err)
		}
	}()

	// HTTP server (API surface will expose engine decisions later)
	srv := server.New(":8000", eng)

	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Block until SIGINT or SIGTERM
	<-stop
	log.Println("shutdown signal received")

	// Begin graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Stop engine
	cancel()

	// Stop HTTP server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}

	log.Println("sentinel exited cleanly")
}
