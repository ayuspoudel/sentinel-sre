package clusterStatusModel

import "time"

type ClusterStatus struct {
	ClusterName              string
	AgentID                  *string
	LastReconcileAt          *time.Time
	LastReconcileDurationMs  *int
	LastReconcileSuccess     *bool
	LastError                *string
	Reachable                *bool
	AuthValid                *bool
	APIServerVersion         *string
	LastSuccessfulConnection *time.Time
	AgentInstalled           *bool
	AgentVersion             *string
	AgentNamespace           *string
	AgentHealthy             *bool
	AgentLastHeartbeat       *time.Time
	UpdatedAt                time.Time
}
