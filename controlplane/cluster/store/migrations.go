package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(ctx context.Context, db *pgxpool.Pool) error {
	query := `CREATE TABLE IF NOT EXISTS clusters (name TEXT PRIMARY KEY, credential_ref TEXT NOT NULL, labels JSONB NOT NULL DEFAULT '{}', created_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`
	_, err := db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("cluster migration failed: %w", err)
	}
	return nil
}
