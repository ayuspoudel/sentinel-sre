package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/adapters"
	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/api"
	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/service"
	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/store"
)

func main() {
	dbURL := mustEnv("DATABASE_URL")
	httpAddr := envOrDefault("HTTP_ADDR", ":9001")
	promURL := mustEnv("PROMETHEUS_BASE_URL")

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

	if err := store.RunMigrations(context.Background(), db); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	policyStore := store.NewPostgresStore(db)

	clusterRegistry := adapters.NewClusterRegistryAdapter(db)
	controllerReader := adapters.NewControllerReaderAdapter(db)
	prometheus := adapters.NewPrometheusAdapter(promURL)

	registryService := service.NewRegistryService(
		policyStore,
		clusterRegistry,
		controllerReader,
		prometheus,
	)

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
