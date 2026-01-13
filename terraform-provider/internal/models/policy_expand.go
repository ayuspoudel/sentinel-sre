package models

import "github.com/ayuspoudel/sentinel-sre/controlplane/policy/spec"

/*
Author: @ayuspoudel
We have a PolicyModel, but sentinel (the system terraform will talk to)
does not know anything about policy model. However, it has its own struct
defined as policySpec under spec package. We will use that package to return
a struct type that matches it, and map our PolicyModel fields into it.
This function will be used when we are about to call Sentinel API with the fields
it needs.
*/
func ExpandPolicySpec(m PolicyModel) *spec.PolicySpec {
	return &spec.PolicySpec{
		Metadata: spec.Metadata{
			Name:        m.Name.ValueString(),
			Owner:       m.Metadata.Owner.ValueString(),
			Environment: m.Metadata.Environment.ValueString(),
		},
		Target: spec.Target{
			Cluster:   m.Target.Cluster.ValueString(),
			Namespace: m.Target.Namespace.ValueString(),
			Service:   m.Target.Service.ValueString(),
		},
		Signals: spec.Signals{
			Traffic: spec.TrafficSignal{
				Query:  m.Signals.Traffic.Query.ValueString(),
				MinRPS: m.Signals.Traffic.MinRPS.ValueFloat64(),
			},
			Errors: spec.ErrorSignal{
				Query: m.Signals.Errors.Query.ValueString(),
			},
			SLO: spec.SLO{
				Objective: m.Signals.SLO.Objective.ValueFloat64(),
				Window:    m.Signals.SLO.Window.ValueString(),
			},
		},
		Policy: spec.Policy{
			Budget: spec.BudgetPolicy{
				FastBurn: spec.BurnWindow{
					Window:    m.Policy.Budget.FastBurn.Window.ValueString(),
					Threshold: m.Policy.Budget.FastBurn.Threshold.ValueFloat64(),
				},
				SlowBurn: spec.BurnWindow{
					Window:    m.Policy.Budget.SlowBurn.Window.ValueString(),
					Threshold: m.Policy.Budget.SlowBurn.Threshold.ValueFloat64(),
				},
			},
		},
	}
}
