package adapters

import (
	"context"
	"database/sql"
)

type ControllerReaderAdapter struct {
	db *sql.DB
}

func NewControllerReaderAdapter(db *sql.DB) *ControllerReaderAdapter {
	return &ControllerReaderAdapter{db: db}
}

func (a *ControllerReaderAdapter) ClusterReachable(ctx context.Context, cluster string) (bool, error) {
	var reachable bool
	err := a.db.QueryRowContext(
		ctx,
		`SELECT api_reachable FROM cluster_status WHERE cluster_name = $1`,
		cluster,
	).Scan(&reachable)
	return reachable, err
}

func (a *ControllerReaderAdapter) AgentInstalled(ctx context.Context, cluster string) (bool, error) {
	var installed bool
	err := a.db.QueryRowContext(
		ctx,
		`SELECT agent_installed FROM cluster_status WHERE cluster_name = $1`,
		cluster,
	).Scan(&installed)
	return installed, err
}

func (a *ControllerReaderAdapter) AgentHealthy(ctx context.Context, cluster string) (bool, error) {
	var healthy bool
	err := a.db.QueryRowContext(
		ctx,
		`SELECT agent_healthy FROM cluster_status WHERE cluster_name = $1`,
		cluster,
	).Scan(&healthy)
	return healthy, err
}

func (a *ControllerReaderAdapter) NamespaceExists(ctx context.Context, cluster, namespace string) (bool, error) {
	var exists bool
	err := a.db.QueryRowContext(
		ctx,
		`SELECT EXISTS (SELECT 1 FROM cluster_namespaces WHERE cluster_name = $1 AND namespace = $2)`,
		cluster,
		namespace,
	).Scan(&exists)
	return exists, err
}
