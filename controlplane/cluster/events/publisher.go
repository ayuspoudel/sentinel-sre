package events

import (
	"context"

	"github.com/ayuspoudel/sentinel-sre/controlplane/cluster/model"
)

type ClusterIntentPublisher interface {
	PublishClusterRegistered(ctx context.Context, c *model.Cluster) error
	PublishClusterDeleted(ctx context.Context, clusterName string) error
}
