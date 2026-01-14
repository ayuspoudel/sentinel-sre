package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/adapters"
	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/api"
	clusterRuntimeEvents "github.com/ayuspoudel/sentinel-sre/controlplane/policy/events/clusterRuntimeEvents"
	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/service"
	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/store/clusterRuntime"
	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/store/policyStatus"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbURL := mustEnv("DATABASE_URL")
	httpAddr := envOrDefault("HTTP_ADDR", ":9001")
	promURL := mustEnv("PROMETHEUS_BASE_URL")
	redisAddr := mustEnv("REDIS_ADDR")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to open postgres: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping postgres: %v", err)
	}

	if err := policyStatus.RunMigrations(ctx, db); err != nil {
		log.Fatalf("policy migration failed: %v", err)
	}
	if err := clusterRuntime.RunMigrations(ctx, db); err != nil {
		log.Fatalf("cluster runtime migration failed: %v", err)
	}

	policyStore := policyStatus.NewPostgresStore(db)
	clusterRuntimeStore := clusterRuntime.NewPostgresStore(db)

	clusterRegistry := adapters.NewClusterRegistryAdapter(db)
	runtimeReader := adapters.NewClusterRuntimeReaderAdapter(clusterRuntimeStore)
	prometheus := adapters.NewPrometheusAdapter(promURL)

	registryService := service.NewRegistryService(
		policyStore,
		clusterRegistry,
		runtimeReader,
		prometheus,
	)

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	clusterRuntimeConsumer := clusterRuntimeEvents.NewConsumer(
		rdb,
		"sentinel.events",
		"policy-registry",
		"policy-registry-1",
		clusterRuntimeStore,
	)

	go clusterRuntimeConsumer.Run(ctx)

	handler := api.NewHandler(registryService)
	server := api.NewServer(httpAddr, handler)

	log.Printf("policy registry listening on %s", httpAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("http server failed: %v", err)
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required environment variable %q is not set", key)
	}
	return v
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
