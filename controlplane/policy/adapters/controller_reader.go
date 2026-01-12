package adapters

import (
	"context"
	"database/sql"
)

/*
	Author: @ayuspoudel
	This adapter lets the policy service observe current cluster reality without talking directly
	to Kubernetes. It relies on state already collected by agent-controller and stored
	in the database.
		1. Cluster Reachable?
		2. Agent Installed?
		3. Agent Healthy?
		4. Namespace Exists?
	These checks are soft signals used to populate policy status, not to block policy
	creation. Keeping this logic as a thin DB reader avoids tight coupling between the policy service
	and cluster APIs while still giving accurate readiness information.
*/

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
