package clusterRegistry

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(ctx context.Context, db *pgxpool.Pool) error {
	query := `
	CREATE TABLE IF NOT EXISTS managed_clusters (
		cluster_name   TEXT PRIMARY KEY,
		credential_ref TEXT NOT NULL,
		labels         JSONB NOT NULL DEFAULT '{}',
		registered_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
		source TEXT NOT NULL
	);
	`
	_, err := db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("managed_clusters migration failed: %w", err)
	}
	return nil
}
