package clusterRuntime

import (
	"context"
	"database/sql"
	"fmt"
)

func RunMigrations(ctx context.Context, db *sql.DB) error {
	query := `
CREATE TABLE IF NOT EXISTS cluster_runtime_status (
	cluster_name TEXT PRIMARY KEY,

	reachable BOOLEAN,
	auth_valid BOOLEAN,

	agent_installed BOOLEAN,
	agent_healthy BOOLEAN,
	agent_version TEXT,
	agent_namespace TEXT,

	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
`
	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("cluster runtime migration failed: %w", err)
	}
	return nil
}
