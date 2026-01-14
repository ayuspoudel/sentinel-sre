package clusterRuntime

import (
	"context"
	"database/sql"
	"time"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) UpsertClusterRuntime(
	ctx context.Context,
	cluster string,
	reachable bool,
	authValid bool,
	agentInstalled bool,
	agentHealthy bool,
	agentVersion string,
	agentNamespace string,
) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO cluster_runtime_status (
		cluster_name,
		reachable,
		auth_valid,
		agent_installed,
		agent_healthy,
		agent_version,
		agent_namespace,
		updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (cluster_name)
		DO UPDATE SET
		reachable = EXCLUDED.reachable,
		auth_valid = EXCLUDED.auth_valid,
		agent_installed = EXCLUDED.agent_installed,
		agent_healthy = EXCLUDED.agent_healthy,
		agent_version = EXCLUDED.agent_version,
		agent_namespace = EXCLUDED.agent_namespace,
		updated_at = EXCLUDED.updated_at
		WHERE
		cluster_runtime_status.reachable      IS DISTINCT FROM EXCLUDED.reachable OR
		cluster_runtime_status.auth_valid      IS DISTINCT FROM EXCLUDED.auth_valid OR
		cluster_runtime_status.agent_installed IS DISTINCT FROM EXCLUDED.agent_installed OR
		cluster_runtime_status.agent_healthy   IS DISTINCT FROM EXCLUDED.agent_healthy OR
		cluster_runtime_status.agent_version   IS DISTINCT FROM EXCLUDED.agent_version OR
		cluster_runtime_status.agent_namespace IS DISTINCT FROM EXCLUDED.agent_namespace;
	`,
		cluster,
		reachable,
		authValid,
		agentInstalled,
		agentHealthy,
		agentVersion,
		agentNamespace,
		time.Now().UTC(),
	)
	return err
}

func (s *PostgresStore) ClusterReachable(ctx context.Context, cluster string) (bool, error) {
	var v bool
	err := s.db.QueryRowContext(ctx,
		`SELECT reachable FROM cluster_runtime_status WHERE cluster_name = $1`,
		cluster,
	).Scan(&v)
	return v, err
}

func (s *PostgresStore) AgentInstalled(ctx context.Context, cluster string) (bool, error) {
	var v bool
	err := s.db.QueryRowContext(ctx,
		`SELECT agent_installed FROM cluster_runtime_status WHERE cluster_name = $1`,
		cluster,
	).Scan(&v)
	return v, err
}

func (s *PostgresStore) AgentHealthy(ctx context.Context, cluster string) (bool, error) {
	var v bool
	err := s.db.QueryRowContext(ctx,
		`SELECT agent_healthy FROM cluster_runtime_status WHERE cluster_name = $1`,
		cluster,
	).Scan(&v)
	return v, err
}

func (s *PostgresStore) AgentNamespace(ctx context.Context, cluster string) (string, error) {
	var ns string
	err := s.db.QueryRowContext(
		ctx,
		`SELECT agent_namespace FROM cluster_runtime_status WHERE cluster_name = $1`,
		cluster,
	).Scan(&ns)
	return ns, err
}
