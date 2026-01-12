package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/spec"
	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/status"
)

/*
Author: @ayuspoudel
This is the implementation of the functions provided by our storage layer.
We have not defined a struct and valdiates user inputs at this level here for simplicity.
*/
type PostgresStore struct{ db *sql.DB }

func NewPostgresStore(db *sql.DB) *PostgresStore { return &PostgresStore{db: db} }

func (s *PostgresStore) UpsertPolicy(ctx context.Context, p *spec.PolicySpec) error {
	specBytes, err := json.Marshal(p)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `INSERT INTO policies (name, spec) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET spec = EXCLUDED.spec, updated_at = now()`, p.Metadata.Name, specBytes); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO policy_status (policy_name, status) VALUES ($1, '{}'::jsonb) ON CONFLICT (policy_name) DO NOTHING`, p.Metadata.Name); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *PostgresStore) GetPolicy(ctx context.Context, name string) (*spec.PolicySpec, error) {
	var specBytes []byte
	err := s.db.QueryRowContext(ctx, `SELECT spec FROM policies WHERE name = $1`, name).Scan(&specBytes)
	if err == sql.ErrNoRows {
		return nil, errors.New("policy not found")
	}
	if err != nil {
		return nil, err
	}
	var p spec.PolicySpec
	if err := json.Unmarshal(specBytes, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *PostgresStore) ListPolicies(ctx context.Context) ([]*spec.PolicySpec, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT spec FROM policies ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*spec.PolicySpec
	for rows.Next() {
		var specBytes []byte
		if err := rows.Scan(&specBytes); err != nil {
			return nil, err
		}
		var p spec.PolicySpec
		if err := json.Unmarshal(specBytes, &p); err != nil {
			return nil, err
		}
		out = append(out, &p)
	}
	return out, nil
}

func (s *PostgresStore) DeletePolicy(ctx context.Context, name string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM policies WHERE name = $1`, name)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("policy not found")
	}
	return nil
}

func (s *PostgresStore) GetStatus(ctx context.Context, policyName string) (*status.PolicyStatus, error) {
	var statusBytes []byte
	err := s.db.QueryRowContext(ctx, `SELECT status FROM policy_status WHERE policy_name = $1`, policyName).Scan(&statusBytes)
	if err == sql.ErrNoRows {
		return nil, errors.New("policy status not found")
	}
	if err != nil {
		return nil, err
	}
	var st status.PolicyStatus
	if err := json.Unmarshal(statusBytes, &st); err != nil {
		return nil, err
	}
	return &st, nil
}

func (s *PostgresStore) UpdateStatus(ctx context.Context, st *status.PolicyStatus) error {
	statusBytes, err := json.Marshal(st)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `UPDATE policy_status SET status = $2, updated_at = now() WHERE policy_name = $1`, st.PolicyName, statusBytes)
	return err
}
