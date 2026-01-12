package adapters

import (
	"context"
	"database/sql"
)

type ClusterRegistryAdapter struct {
	db *sql.DB
}

func NewClusterRegistryAdapter(db *sql.DB) *ClusterRegistryAdapter {
	return &ClusterRegistryAdapter{db: db}
}

func (a *ClusterRegistryAdapter) ClusterExists(ctx context.Context, name string) (bool, error) {
	var exists bool
	err := a.db.QueryRowContext(
		ctx,
		`SELECT EXISTS (SELECT 1 FROM clusters WHERE name = $1)`,
		name,
	).Scan(&exists)
	return exists, err
}
