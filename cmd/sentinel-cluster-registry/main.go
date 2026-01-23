package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ayuspoudel/sentinel-sre/controlplane/cluster/api"
	"github.com/ayuspoudel/sentinel-sre/controlplane/cluster/db"
	redisPublisher "github.com/ayuspoudel/sentinel-sre/controlplane/cluster/events/redis"
	"github.com/ayuspoudel/sentinel-sre/controlplane/cluster/kube"
	"github.com/ayuspoudel/sentinel-sre/controlplane/cluster/redis"
	"github.com/ayuspoudel/sentinel-sre/controlplane/cluster/service"
	"github.com/ayuspoudel/sentinel-sre/controlplane/cluster/store"
	"github.com/ayuspoudel/sentinel-sre/pkg/env"
)

func main() {
	dbURL := os.Getenv("REGISTRY_DB_URL")
	if dbURL == "" {
		log.Fatal("REGISTRY_DB_URL is required")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	dbCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	db, err := db.NewDB(dbCtx, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	migCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err = store.RunMigrations(migCtx, db)
	if err != nil {
		log.Fatal(err)
	}

	store := store.NewPgxStore(db)
	svc := service.NewService(store)
	redis := redis.New()
	redisStream := env.MustEnv("REDIS_STREAM", false)
	var publisher redisPublisher.RedisPublisher
	if redisStream != "" {
		publisher = *redisPublisher.NewRedisPublisher(redis, redisStream)
	}
	kubeClient, err := kube.NewKube()
	if err != nil {
		log.Fatal(err)
	}

	handler := api.NewHandler(svc, &publisher, kubeClient)
	server := api.NewServer(":8080", handler)

	err = server.Run(ctx)
	if err != nil {
		log.Fatal(err)
	}

}
