package clusterRegistryModel

import "time"

/*
	Author: @ayuspoudel
	This model represents cluster intent as observed by the agent controller.
	It is populated via cluster.registered events emitted by cluster registry.
	This is NOT runtime state.
*/

type ManagedCluster struct {
	ClusterName   string            `json:"cluster_name" db:"cluster_name"`
	CredentialRef string            `json:"credential_ref" db:"credential_ref"`
	Labels        map[string]string `json:"labels" db:"labels"`
	RegisteredAt  time.Time         `json:"registered_at" db:"registered_at"`
	Source        string            `json:"source" db:"source"`
}
