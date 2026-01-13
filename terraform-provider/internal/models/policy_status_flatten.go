package models

import (
	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/status"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func FlattenPolicyStatus(s *status.PolicyStatus) PolicyStatusModel {
	m := PolicyStatusModel{
		Name: types.StringValue(s.PolicyName),

		ClusterExists:    types.BoolValue(s.ClusterExists),
		ClusterReachable: types.BoolValue(s.ClusterReachable),
		NamespaceExists:  types.BoolValue(s.NamespaceExists),

		AgentInstalled: types.BoolValue(s.AgentInstalled),
		AgentHealthy:   types.BoolValue(s.AgentHealthy),

		PrometheusReachable: types.BoolValue(s.PrometheusReachable),
		QueriesValid:        types.BoolValue(s.QueriesValid),

		LastValidatedAt: types.StringValue(s.LastValidatedAt.Format("2006-01-02T15:04:05Z")),
	}

	if s.LastError != nil {
		m.LastError = types.StringValue(*s.LastError)
	} else {
		m.LastError = types.StringNull()
	}

	return m
}
