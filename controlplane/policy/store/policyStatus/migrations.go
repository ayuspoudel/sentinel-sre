package policyStatus

import (
	"context"
	"database/sql"
	"fmt"
)

/*
	Author: @ayuspoudel
	This defines the DB schema policy registry will create at startup time
	If the db table does not already exist in the db
*/

func RunMigrations(ctx context.Context, db *sql.DB) error {
	query := `
CREATE TABLE IF NOT EXISTS policies (
    name TEXT PRIMARY KEY,
    spec JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS policy_status (
    policy_name TEXT PRIMARY KEY,
    status JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_policy
        FOREIGN KEY(policy_name)
        REFERENCES policies(name)
        ON DELETE CASCADE
);
`
	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("policy migration failed: %w", err)
	}
	return nil
}
