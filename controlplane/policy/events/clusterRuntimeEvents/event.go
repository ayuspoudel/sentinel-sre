package clusterRuntimeStatusEvents

import "time"

type ClusterStatusEvent struct {
	SchemaVersion int       `json:"schema_version"`
	Type          string    `json:"type"`
	Source        string    `json:"source"`
	ClusterName   string    `json:"cluster_name"`
	Timestamp     time.Time `json:"timestamp"`
	Status        struct {
		Reachable      bool   `json:"reachable"`
		AuthValid      bool   `json:"auth_valid"`
		AgentInstalled bool   `json:"agent_installed"`
		AgentHealthy   bool   `json:"agent_healthy"`
		AgentVersion   string `json:"agent_version"`
		AgentNamespace string `json:"agent_namespace"`
	} `json:"status"`
}
