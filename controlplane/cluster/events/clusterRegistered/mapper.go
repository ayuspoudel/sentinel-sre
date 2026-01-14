package clusterRegistered

import (
	"github.com/ayuspoudel/sentinel-sre/controlplane/cluster/model"
)

/*
	Author: @ayuspoudel
	Map internal cluster model into publishable cluster registered event.
*/

func FromCluster(c *model.Cluster) *ClusterPayload {
	return &ClusterPayload{
		Name:          c.Name,
		CredentialRef: c.CredentialRef,
		Labels:        c.Labels,
	}
}
