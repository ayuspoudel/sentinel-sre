package clusterStatusModel

import "time"

func (c *ClusterStatus) Merge(old, new *ClusterStatus) *ClusterStatus {
	if old == nil {
		return new
	}

	merged := *old

	mergePtr(&merged.AgentID, new.AgentID)
	mergePtr(&merged.LastReconcileAt, new.LastReconcileAt)
	mergePtr(&merged.LastReconcileDurationMs, new.LastReconcileDurationMs)
	mergePtr(&merged.LastReconcileSuccess, new.LastReconcileSuccess)
	mergePtr(&merged.LastError, new.LastError)

	mergePtr(&merged.Reachable, new.Reachable)
	mergePtr(&merged.AuthValid, new.AuthValid)
	mergePtr(&merged.APIServerVersion, new.APIServerVersion)
	mergePtr(&merged.LastSuccessfulConnection, new.LastSuccessfulConnection)

	mergePtr(&merged.AgentInstalled, new.AgentInstalled)
	mergePtr(&merged.AgentVersion, new.AgentVersion)
	mergePtr(&merged.AgentNamespace, new.AgentNamespace)
	mergePtr(&merged.AgentHealthy, new.AgentHealthy)
	mergePtr(&merged.AgentLastHeartbeat, new.AgentLastHeartbeat)

	merged.UpdatedAt = time.Now()
	return &merged
}

func mergePtr[T any](dst **T, src *T) {
	if src != nil {
		*dst = src
	}
}
