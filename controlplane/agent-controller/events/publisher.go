package events

import (
	"context"

	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/status"
)

type ClusterStatusPublisher interface {
	PublishClusterStatus(ctx context.Context, st *status.ClusterStatus) error
}
