package events

import "github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/status"

/*
	Author: @ayuspoudel
	We have defined our cluster status event in event.go. Now we need to map our internal controller's
	cluster status into the publishable event we have defined.
*/

func FromClusterStatus(st *status.ClusterStatus) ClusterRuntimeStatus {
	return ClusterRuntimeStatus{
		Reachable:      st.Reachable,
		AuthValid:      st.AuthValid,
		AgentInstalled: st.AgentInstalled,
		AgentHealthy:   st.AgentHealthy,
		AgentVersion:   st.AgentVersion,
		AgentNamespace: st.AgentNamespace,
	}
}
