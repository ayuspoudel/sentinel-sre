package events

import (
	"context"

	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/status"
)

type ClusterStatusPublisher interface {
	PublishClusterStatus(ctx context.Context, st *status.ClusterStatus) error
}
