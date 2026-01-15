package cluster

import "context"

type Store interface {
	Create(ctx context.Context, cluster *Cluster) error
	Get(ctx context.Context, name string) (*Cluster, error)
	List(ctx context.Context) ([]*Cluster, error)
	Delete(ctx context.Context, name string) error
}
