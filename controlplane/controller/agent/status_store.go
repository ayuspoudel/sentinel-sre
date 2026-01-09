package agent

import "context"

type Store interface {
	Upsert(ctx context.Context, st *ClusterStatus) error
	Get(ctx context.Context, clusterName string) (*ClusterStatus, error)
}
