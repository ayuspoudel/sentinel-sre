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

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	registryURL := mustEnv("REGISTRY_URL")
	controlPlaneURL := mustEnv("CONTROL_PLANE_URL")
	postgresURL := mustEnv("AGENT_DB_URL")

	reconcileInterval := envDuration("RECONCILE_INTERVAL", 10*time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	registry := registryClient.New(registryURL)
	kubeClient, err := agent.NewKubeClient()
	if err != nil {
		log.Fatalf("failed to init kube client: %v", err)
	}
	db, err := pgxpool.New(ctx, postgresURL)
	if err != nil {
		log.Fatalf("failed to connect postgres: %v", err)
	}
	if err := agent.RunMigrations(ctx, db); err != nil {
		log.Fatalf("db migrations failed: %v", err)
	}

	store := agent.NewStatusStore(db)

	installer := agent.NewHelmInstaller(
		"sentinel-agent",
		"https://ayuspoudel.github.io/sentinel-sre",
		"sentinel-sre",
	)

	controller := agent.NewController(
		registry,
		kubeClient,
		store,
		reconcileInterval,
		controlPlaneURL,
		installer,
	)

	go controller.Run(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutdown signal received")
	cancel()
	time.Sleep(1 * time.Second)
	log.Println("sentinel agent controller exited cleanly")
}

func mustEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("environment variable %s is required", name)
	}
	return value
}

func envDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
