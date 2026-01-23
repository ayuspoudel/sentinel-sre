package clusterDeleted

import "time"

type ClusterDeletedEvent struct {
	SchemaVersion int       `json:"schema_version"`
	Type          string    `json:"type"`
	Source        string    `json:"source"`
	TimeStamp     time.Time `json:"timestamp"`
	ClusterName   string    `json:"cluster_name"`
}

func NewClusterDeletedEvent(clusterName string) *ClusterDeletedEvent {
	return &ClusterDeletedEvent{
		SchemaVersion: 1,
		Type:          "cluster.deleted",
		Source:        "cluster-registry",
		TimeStamp:     time.Now(),
		ClusterName:   clusterName,
	}
}
