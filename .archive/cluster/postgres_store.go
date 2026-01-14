package cluster

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgxStore struct {
	db *pgxpool.Pool
}

func NewPgxStore(db *pgxpool.Pool) *PgxStore {
	return &PgxStore{db: db}
}

func (s *PgxStore) Create(ctx context.Context, c *Cluster) error {
	query := `INSERT INTO clusters (name, credential_ref, labels) VALUES ($1, $2, $3) ON CONFLICT (name) DO NOTHING`
	_, err := s.db.Exec(ctx, query, c.Name, c.CredentialRef, c.Labels)
	return err
}

func (s *PgxStore) Get(ctx context.Context, name string) (*Cluster, error) {
	query := `SELECT name, credential_ref, labels, created_at FROM clusters WHERE name=$1`
	row := s.db.QueryRow(ctx, query, name)
	var c Cluster
	err := row.Scan(&c.Name, &c.CredentialRef, &c.Labels, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (s *PgxStore) List(ctx context.Context) ([]*Cluster, error) {
	query := `SELECT name, credential_ref, labels, created_at FROM clusters ORDER BY created_at`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var clusters []*Cluster
	for rows.Next() {
		var c Cluster
		err := rows.Scan(&c.Name, &c.CredentialRef, &c.Labels, &c.CreatedAt)
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, &c)
	}
	rowsErr := rows.Err()
	if rowsErr != nil {
		return nil, rowsErr

	}
	return clusters, nil
}

func (s *PgxStore) Delete(ctx context.Context, name string) error {
	query := `DELETE FROM clusters WHERE name=$1`
	_, err := s.db.Exec(ctx, query, name)
	return err
}
