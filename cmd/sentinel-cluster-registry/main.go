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
	"github.com/ayuspoudel/sentinel-sre/controlplane/cluster/service"
	"github.com/ayuspoudel/sentinel-sre/controlplane/cluster/store"
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
	handler := api.NewHandler(svc)
	server := api.NewServer(":8080", handler)

	err = server.Run(ctx)
	if err != nil {
		log.Fatal(err)
	}

}
