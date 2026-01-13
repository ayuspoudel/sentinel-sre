package models

import "github.com/hashicorp/terraform-plugin-framework/types"

func testPolicyModel() PolicyModel {
	return PolicyModel{
		Name: types.StringValue("checkout"),
		Metadata: MetadataModel{
			Owner:       types.StringValue("team-a"),
			Environment: types.StringValue("prod"),
		},
		Target: TargetModel{
			Cluster:   types.StringValue("sreCluster"),
			Namespace: types.StringValue("default"),
			Service:   types.StringValue("checkout"),
		},
		Signals: SignalsModel{
			Traffic: TrafficSignalModel{
				Query:  types.StringValue("sum(rate(http_requests_total[1m]))"),
				MinRPS: types.Float64Value(10),
			},
			Errors: ErrorSignalModel{
				Query: types.StringValue("sum(rate(http_requests_errors_total[1m]))"),
			},
			SLO: SLOModel{
				Objective: types.Float64Value(99.9),
				Window:    types.StringValue("720h"),
			},
		},
		Policy: PolicyRuleModel{
			Budget: BudgetPolicyModel{
				FastBurn: BurnWindowModel{
					Window:    types.StringValue("5m"),
					Threshold: types.Float64Value(2),
				},
				SlowBurn: BurnWindowModel{
					Window:    types.StringValue("1h"),
					Threshold: types.Float64Value(1),
				},
			},
		},
	}
}
