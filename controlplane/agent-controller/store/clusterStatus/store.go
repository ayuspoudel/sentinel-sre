package clusterStatus

import (
	"context"

	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/models/clusterStatusModel"
)

type Store interface {
	Upsert(ctx context.Context, st *clusterStatusModel.ClusterStatus) error
	Get(ctx context.Context, clusterName string) (*clusterStatusModel.ClusterStatus, error)
}
