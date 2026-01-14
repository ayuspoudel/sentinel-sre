package adapters

import (
	"context"

	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/store/clusterRuntime"
)

type ClusterRuntimeReaderAdapter struct {
	store clusterRuntime.Store
}

func NewClusterRuntimeReaderAdapter(store clusterRuntime.Store) *ClusterRuntimeReaderAdapter {
	return &ClusterRuntimeReaderAdapter{store: store}
}

func (a *ClusterRuntimeReaderAdapter) ClusterReachable(ctx context.Context, cluster string) (bool, error) {
	return a.store.ClusterReachable(ctx, cluster)
}

func (a *ClusterRuntimeReaderAdapter) AgentInstalled(ctx context.Context, cluster string) (bool, error) {
	return a.store.AgentInstalled(ctx, cluster)
}

func (a *ClusterRuntimeReaderAdapter) AgentHealthy(ctx context.Context, cluster string) (bool, error) {
	return a.store.AgentHealthy(ctx, cluster)
}

func (a *ClusterRuntimeReaderAdapter) NamespaceExists(ctx context.Context, cluster, _ string) (bool, error) {
	// NamespaceExists now means: Sentinel agent namespace exists.
	ns, err := a.store.AgentNamespace(ctx, cluster)
	if err != nil {
		return false, err
	}
	return ns != "", nil
}
