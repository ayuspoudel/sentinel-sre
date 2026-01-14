package events

import "time"

/*
	Author: @ayuspoudel
	Other microservices within this control plane will need cluster status. Policy registry might need if cluster is reachable from
	control plane or not. Instead of running evaluation loop in policy registry and other consumers, we will publish an event whenever
	the cluster status changes as evaluated by agent controller.
	This defines what the event will have, the structure of the data in the event.
*/

type ClusterStatusEvent struct {
	SchemaVersion int                  `json:"schema_version"`
	Type          string               `json:"type"`
	Source        string               `json:"source"`
	ClusterName   string               `json:"cluster_name"`
	TimeStamp     time.Time            `json:"timestamp"`
	Status        ClusterRuntimeStatus `json:"status"`
}

type ClusterRuntimeStatus struct {
	Reachable *bool `json:"reachable,omitempty"`
	AuthValid *bool `json:"auth_valid,omitempty"`

	AgentInstalled *bool   `json:"agent_installed,omitempty"`
	AgentHealthy   *bool   `json:"agent_healthy,omitempty"`
	AgentVersion   *string `json:"agent_version,omitempty"`
	AgentNamespace *string `json:"agent_namespace,omitempty"`
}

func NewClusterStatusEvent(clusterName string, status ClusterRuntimeStatus) *ClusterStatusEvent {
	return &ClusterStatusEvent{
		SchemaVersion: 1,
		Type:          "agent.cluster_status.updated",
		Source:        "agent-controller",
		ClusterName:   clusterName,
		TimeStamp:     time.Now(),
		Status:        status,
	}

}
