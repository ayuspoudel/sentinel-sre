package agent

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type StatusStore struct {
	db *pgxpool.Pool
}

func NewStatusStore(db *pgxpool.Pool) *StatusStore {
	return &StatusStore{db: db}
}

func (s *StatusStore) Upsert(ctx context.Context, st *ClusterStatus) error {
	query := `INSERT INTO cluster_status (
		cluster_name,
		last_reconcile_at,
		last_reconcile_duration_ms,
		last_reconcile_success,
		last_error,
		reachable,
		auth_valid,
		api_server_version,
		last_successful_connection,
		agent_installed,
		agent_version,
		agent_namespace,
		agent_healthy,
		agent_last_heartbeat,
		updated_at
	) VALUES (
		$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15
	) ON CONFLICT (cluster_name) DO UPDATE SET
		last_reconcile_at=EXCLUDED.last_reconcile_at,
		last_reconcile_duration_ms=EXCLUDED.last_reconcile_duration_ms,
		last_reconcile_success=EXCLUDED.last_reconcile_success,
		last_error=EXCLUDED.last_error,
		reachable=EXCLUDED.reachable,
		auth_valid=EXCLUDED.auth_valid,
		api_server_version=EXCLUDED.api_server_version,
		last_successful_connection=EXCLUDED.last_successful_connection,
		agent_installed=EXCLUDED.agent_installed,
		agent_version=EXCLUDED.agent_version,
		agent_namespace=EXCLUDED.agent_namespace,
		agent_healthy=EXCLUDED.agent_healthy,
		agent_last_heartbeat=EXCLUDED.agent_last_heartbeat,
		updated_at=EXCLUDED.updated_at`
	_, err := s.db.Exec(
		ctx,
		query,
		st.ClusterName,
		st.LastReconcileAt,
		st.LastReconcileDurationMs,
		st.LastReconcileSuccess,
		st.LastError,
		st.Reachable,
		st.AuthValid,
		st.APIServerVersion,
		st.LastSuccessfulConnection,
		st.AgentInstalled,
		st.AgentVersion,
		st.AgentNamespace,
		st.AgentHealthy,
		st.AgentLastHeartbeat,
		time.Now(),
	)
	return err
}

func (s *StatusStore) Get(ctx context.Context, clusterName string) (*ClusterStatus, error) {
	query := `SELECT
		cluster_name,
		last_reconcile_at,
		last_reconcile_duration_ms,
		last_reconcile_success,
		last_error,
		reachable,
		auth_valid,
		api_server_version,
		last_successful_connection,
		agent_installed,
		agent_version,
		agent_namespace,
		agent_healthy,
		agent_last_heartbeat,
		updated_at
	FROM cluster_status WHERE cluster_name=$1`
	row := s.db.QueryRow(ctx, query, clusterName)
	var st ClusterStatus
	err := row.Scan(
		&st.ClusterName,
		&st.LastReconcileAt,
		&st.LastReconcileDurationMs,
		&st.LastReconcileSuccess,
		&st.LastError,
		&st.Reachable,
		&st.AuthValid,
		&st.APIServerVersion,
		&st.LastSuccessfulConnection,
		&st.AgentInstalled,
		&st.AgentVersion,
		&st.AgentNamespace,
		&st.AgentHealthy,
		&st.AgentLastHeartbeat,
		&st.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &st, nil
}
