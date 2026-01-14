package clusterRegistered

import (
	"context"

	"github.com/ayuspoudel/sentinel-sre/controlplane/cluster/model"
)

type ClusterRegisteredPublisher interface {
	PublishClusterRegistered(ctx context.Context, c *model.Cluster) error
}
