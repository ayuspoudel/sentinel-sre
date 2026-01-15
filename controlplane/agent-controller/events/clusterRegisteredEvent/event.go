package clusterRegisteredEvent

import "time"

type ClusterRegisteredEvent struct {
	SchemaVersion int            `json:"schema_version"`
	Type          string         `json:"type"`
	Source        string         `json:"source"`
	Timestamp     time.Time      `json:"timestamp"`
	Cluster       ClusterPayload `json:"cluster"`
}

type ClusterPayload struct {
	Name          string            `json:"name"`
	CredentialRef string            `json:"credential_ref"`
	Labels        map[string]string `json:"labels"`
}
