package adapters

import (
	"context"
	"database/sql"
)

/*
	Author: @ayuspoudel
	This is a cluster registry adaptor. It assumes that cluster registry and policy registry will share the same DB
	and cluster registry's table is named clusers.
	We want to get information from cluster registry about reachability of cluster, health of agent and so on.

*/

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
