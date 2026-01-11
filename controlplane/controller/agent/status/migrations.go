package status

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(ctx context.Context, db *pgxpool.Pool) error {
	query := `CREATE TABLE IF NOT EXISTS cluster_status (
		cluster_name TEXT PRIMARY KEY,
		last_reconcile_at TIMESTAMPTZ,
		last_reconcile_duration_ms INTEGER,
		last_reconcile_success BOOLEAN,
		last_error TEXT,
		reachable BOOLEAN,
		auth_valid BOOLEAN,
		api_server_version TEXT,
		last_successful_connection TIMESTAMPTZ,
		agent_installed BOOLEAN,
		agent_version TEXT,
		agent_namespace TEXT,
		agent_healthy BOOLEAN,
		agent_last_heartbeat TIMESTAMPTZ,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`
	_, err := db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("cluster_status migration failed: %w", err)
	}
	return nil
}
