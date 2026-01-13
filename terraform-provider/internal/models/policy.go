package models

import "github.com/hashicorp/terraform-plugin-framework/types"

/*
	Author: @ayuspoudel
	Doc: https://developer.hashicorp.com/terraform/plugin/framework/
	This file contains all the models required for policy resource. Sentinel-policy-registry
	defines a policy manifest with fields like name, metadata, target, signals and policy.
	Each of these fields are defined as a struct below with appropriate types from terraform-plugin-framework/types.
	Reference: https://github.com/ayuspoudel/sentinel-sre/blob/develop/controlplane/policy/spec/spec.go
*/

type PolicyModel struct {
	Name     types.String    `tfsdk:"name"`
	Metadata MetadataModel   `tfsdk:"metadata"`
	Target   TargetModel     `tfsdk:"target"`
	Signals  SignalsModel    `tfsdk:"signals"`
	Policy   PolicyRuleModel `tfsdk:"policy"`
}

type MetadataModel struct {
	Owner       types.String `tfsdk:"owner"`
	Environment types.String `tfsdk:"environment"`
}

type TargetModel struct {
	Cluster   types.String `tfsdk:"cluster"`
	Namespace types.String `tfsdk:"namespace"`
	Service   types.String `tfsdk:"service"`
}

type SignalsModel struct {
	Traffic TrafficSignalModel `tfsdk:"traffic"`
	Errors  ErrorSignalModel   `tfsdk:"errors"`
	SLO     SLOModel           `tfsdk:"slo"`
}

type TrafficSignalModel struct {
	Query  types.String  `tfsdk:"query"`
	MinRPS types.Float64 `tfsdk:"min_rps"`
}

type ErrorSignalModel struct {
	Query types.String `tfsdk:"query"`
}

type SLOModel struct {
	Objective types.Float64 `tfsdk:"objective"`
	Window    types.String  `tfsdk:"window"`
}

type PolicyRuleModel struct {
	Budget BudgetPolicyModel `tfsdk:"budget"`
}

type BudgetPolicyModel struct {
	FastBurn BurnWindowModel `tfsdk:"fast_burn"`
	SlowBurn BurnWindowModel `tfsdk:"slow_burn"`
}

type BurnWindowModel struct {
	Window    types.String  `tfsdk:"window"`
	Threshold types.Float64 `tfsdk:"threshold"`
}
