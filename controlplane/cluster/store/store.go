package store

import (
	"context"

	"github.com/ayuspoudel/sentinel-sre/controlplane/cluster/model"
)

type Store interface {
	Create(ctx context.Context, cluster *model.Cluster) error
	Get(ctx context.Context, name string) (*model.Cluster, error)
	List(ctx context.Context) ([]*model.Cluster, error)
	Delete(ctx context.Context, name string) error
}
