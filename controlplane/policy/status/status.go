package status

import "time"

/*
	@ayuspoudel
	spec/spec.go define the specifications of what we want to recieve as input
	This file define the status structure of policy after evaluation
*/

type PolicyStatus struct {
	PolicyName string `json:"policy_name"`

	// Environment validation
	ClusterExists    bool `json:"cluster_exists"`
	ClusterReachable bool `json:"cluster_reachable"`
	NamespaceExists  bool `json:"namespace_exists"`

	// Sentinel agent state
	AgentInstalled bool `json:"agent_installed"`
	AgentHealthy   bool `json:"agent_healthy"`

	// Metrics readiness
	PrometheusReachable bool `json:"prometheus_reachable"`
	QueriesValid        bool `json:"queries_valid"`

	// Metadata
	LastValidatedAt time.Time `json:"last_validated_at"`
	LastError       *string   `json:"last_error,omitempty"`
}
