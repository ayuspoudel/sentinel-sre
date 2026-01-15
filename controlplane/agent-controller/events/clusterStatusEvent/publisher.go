package clusterStatusEvent

import (
	"context"

	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/models/clusterStatusModel"
)

type ClusterStatusPublisher interface {
	PublishClusterStatus(ctx context.Context, st *clusterStatusModel.ClusterStatus) error
}
