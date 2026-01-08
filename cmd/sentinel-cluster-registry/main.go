package main

import (
	"context"
	"log"
	"os"

	"github.com/ayuspoudel/sentinel-sre/controlplane/cluster"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()

	db, err := cluster.NewDB(ctx, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = cluster.RunMigrations(ctx, db)
	if err != nil {
		log.Fatal(err)
	}

	store := cluster.NewPgxStore(db)
	handler := cluster.NewHandler(store)
	server := cluster.NewServer(":8080", handler)

	log.Fatal(server.Run())
}
