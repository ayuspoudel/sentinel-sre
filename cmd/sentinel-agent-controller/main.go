package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/controller"
	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/events"
	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/heartbeat"
	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/install"
	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/kube"
	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/status"
	"github.com/ayuspoudel/sentinel-sre/controlplane/registryClient"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	registryURL := mustEnv("REGISTRY_URL", false)
	controlPlaneURL := mustEnv("CONTROL_PLANE_URL", false)
	postgresURL := mustEnv("AGENT_DB_URL", false)
	redisURL := mustEnv("REDIS_URL", false)
	redisPassword := mustEnv("REDIS_PASSWORD", true)
	redisStream := mustEnv("REDIS_STREAM", false)
	redisDB := 0

	if v := mustEnv("REDIS_DB", false); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			redisDB = parsed
		}
	}
	redisClient := events.NewRedisClient(events.RedisConfig{Addr: redisURL, Password: redisPassword, DB: redisDB})
	if err := events.PingRedis(ctx, redisClient); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	publisher := events.NewRedisPublisher(redisClient, redisStream)

	reconcileInterval := envDuration("RECONCILE_INTERVAL", 10*time.Second)

	registry := registryClient.New(registryURL)
	kubeClient, err := kube.NewKubeClient()
	if err != nil {
		log.Fatalf("failed to init kube client: %v", err)
	}
	db, err := pgxpool.New(ctx, postgresURL)
	if err != nil {
		log.Fatalf("failed to connect postgres: %v", err)
	}
	if err := status.RunMigrations(ctx, db); err != nil {
		log.Fatalf("db migrations failed: %v", err)
	}

	store := status.NewStatusStore(db)

	installer := install.NewHelmInstaller(
		"sentinel-agent",
		"https://ayuspoudel.github.io/sentinel-sre",
		"sentinel-sre",
	)

	controller := controller.NewController(registry, kubeClient, store, reconcileInterval, controlPlaneURL, installer, publisher)

	heartbeatAddr := os.Getenv("HEARTBEAT_BIND_ADDR")
	if heartbeatAddr == "" {
		heartbeatAddr = ":9000"
	}

	heartbeatServer := heartbeat.NewServer(
		heartbeatAddr,
		store,
	)

	go func() {
		if err := heartbeatServer.Start(ctx); err != nil {
			log.Printf("heartbeat server exited: %v", err)
		}
	}()

	go controller.Run(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutdown signal received")
	cancel()
	time.Sleep(1 * time.Second)
	log.Println("sentinel agent controller exited cleanly")
}

func mustEnv(name string, omitempty bool) string {
	value := os.Getenv(name)
	if value == "" && omitempty == false {
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
