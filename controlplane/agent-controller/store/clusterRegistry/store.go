package clusterRegistry

import (
	"context"

	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/models/clusterRegistryModel"
)

type Store interface {
	Insert(ctx context.Context, c *clusterRegistryModel.ManagedCluster) error
	Get(ctx context.Context, clusterName string) (*clusterRegistryModel.ManagedCluster, error)
	List(ctx context.Context) ([]*clusterRegistryModel.ManagedCluster, error)
	Delete(ctx context.Context, clusterName string) error
}
