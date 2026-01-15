package clusterRegistry

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/models/clusterRegistryModel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	db *pgxpool.Pool
}

func NewPostgresStore(db *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) Insert(ctx context.Context, c *clusterRegistryModel.ManagedCluster) error {
	labelsBytes, err := json.Marshal(c.Labels)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(ctx, `
		INSERT INTO managed_clusters (
			cluster_name,
			credential_ref,
			labels,
			registered_at,
			source
		)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (cluster_name)
		DO UPDATE SET
			credential_ref = EXCLUDED.credential_ref,
			labels = EXCLUDED.labels,
			registered_at = EXCLUDED.registered_at,
			source = EXCLUDED.source
	`,
		c.ClusterName,
		c.CredentialRef,
		labelsBytes,
		c.RegisteredAt,
		c.Source,
	)
	return err
}

func (s *PostgresStore) Get(ctx context.Context, clusterName string) (*clusterRegistryModel.ManagedCluster, error) {
	row := s.db.QueryRow(ctx, `
		SELECT cluster_name, credential_ref, labels, registered_at, source
		FROM managed_clusters
		WHERE cluster_name = $1
	`, clusterName)

	var c clusterRegistryModel.ManagedCluster
	var labelsBytes []byte

	err := row.Scan(&c.ClusterName, &c.CredentialRef, &labelsBytes, &c.RegisteredAt, &c.Source)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(labelsBytes, &c.Labels); err != nil {
		return nil, err
	}

	return &c, nil
}

func (s *PostgresStore) List(ctx context.Context) ([]*clusterRegistryModel.ManagedCluster, error) {
	rows, err := s.db.Query(ctx, `
		SELECT cluster_name, credential_ref, labels, registered_at, source
		FROM managed_clusters
		ORDER BY registered_at
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*clusterRegistryModel.ManagedCluster

	for rows.Next() {
		var c clusterRegistryModel.ManagedCluster
		var labelsBytes []byte

		if err := rows.Scan(&c.ClusterName, &c.CredentialRef, &labelsBytes, &c.RegisteredAt, &c.Source); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(labelsBytes, &c.Labels); err != nil {
			return nil, err
		}

		out = append(out, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (s *PostgresStore) Delete(ctx context.Context, clusterName string) error {
	_, err := s.db.Exec(ctx, `
		DELETE FROM managed_clusters WHERE cluster_name = $1
	`, clusterName)
	return err
}
