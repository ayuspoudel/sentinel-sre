package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type PolicyStatusModel struct {
	Name types.String `tfsdk:"name"`

	ClusterExists    types.Bool `tfsdk:"cluster_exists"`
	ClusterReachable types.Bool `tfsdk:"cluster_reachable"`
	NamespaceExists  types.Bool `tfsdk:"namespace_exists"`

	AgentInstalled types.Bool `tfsdk:"agent_installed"`
	AgentHealthy   types.Bool `tfsdk:"agent_healthy"`

	PrometheusReachable types.Bool `tfsdk:"prometheus_reachable"`
	QueriesValid        types.Bool `tfsdk:"queries_valid"`

	LastValidatedAt types.String `tfsdk:"last_validated_at"`
	LastError       types.String `tfsdk:"last_error"`
}
