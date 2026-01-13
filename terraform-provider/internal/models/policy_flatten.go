package models

import (
	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/spec"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func FlattenPolicySpec(p *spec.PolicySpec) PolicyModel {
	return PolicyModel{
		Name: types.StringValue(p.Metadata.Name),
		Metadata: MetadataModel{
			Owner:       types.StringValue(p.Metadata.Owner),
			Environment: types.StringValue(p.Metadata.Environment),
		},
		Target: TargetModel{
			Cluster:   types.StringValue(p.Target.Cluster),
			Namespace: types.StringValue(p.Target.Namespace),
			Service:   types.StringValue(p.Target.Service),
		},
		Signals: SignalsModel{
			Traffic: TrafficSignalModel{
				Query:  types.StringValue(p.Signals.Traffic.Query),
				MinRPS: types.Float64Value(p.Signals.Traffic.MinRPS),
			},
			Errors: ErrorSignalModel{
				Query: types.StringValue(p.Signals.Errors.Query),
			},
			SLO: SLOModel{
				Objective: types.Float64Value(p.Signals.SLO.Objective),
				Window:    types.StringValue(p.Signals.SLO.Window),
			},
		},
		Policy: PolicyRuleModel{
			Budget: BudgetPolicyModel{
				FastBurn: BurnWindowModel{
					Window:    types.StringValue(p.Policy.Budget.FastBurn.Window),
					Threshold: types.Float64Value(p.Policy.Budget.FastBurn.Threshold),
				},
				SlowBurn: BurnWindowModel{
					Window:    types.StringValue(p.Policy.Budget.SlowBurn.Window),
					Threshold: types.Float64Value(p.Policy.Budget.SlowBurn.Threshold),
				},
			},
		},
	}
}
