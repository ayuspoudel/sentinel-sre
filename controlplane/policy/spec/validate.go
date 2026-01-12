package spec

import (
	"fmt"
	"time"
)

// This file is a sanity check for all fields

func Validate(p *PolicySpec) error {
	if p.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if p.Target.Cluster == "" {
		return fmt.Errorf("target.cluster is required")
	}
	if p.Target.Namespace == "" {
		return fmt.Errorf("target.namespace is required")
	}

	if p.Signals.Traffic.MinRPS < 0 {
		return fmt.Errorf("signals.traffic.minRPS must be >= 0")
	}

	if p.Signals.SLO.Objective <= 0 || p.Signals.SLO.Objective >= 100 {
		return fmt.Errorf("signals.slo.objective must be between 0 and 100")
	}

	if p.Policy.Budget.FastBurn.Threshold <= 0 {
		return fmt.Errorf("policy.budget.fastBurn.threshold must be > 0")
	}

	if p.Policy.Budget.SlowBurn.Threshold <= 0 {
		return fmt.Errorf("policy.budget.slowBurn.threshold must be > 0")
	}

	if p.Policy.Budget.FastBurn.Threshold <= p.Policy.Budget.SlowBurn.Threshold {
		return fmt.Errorf("fast burn threshold must be greater than slow burn threshold")
	}

	sloWindow, err := time.ParseDuration(p.Signals.SLO.Window)
	if err != nil {
		return fmt.Errorf("signals.slo.window is invalid: %w", err)
	}

	fastWindow, err := time.ParseDuration(p.Policy.Budget.FastBurn.Window)
	if err != nil {
		return fmt.Errorf("policy.budget.fastBurn.window is invalid: %w", err)
	}

	slowWindow, err := time.ParseDuration(p.Policy.Budget.SlowBurn.Window)
	if err != nil {
		return fmt.Errorf("policy.budget.slowBurn.window is invalid: %w", err)
	}

	if fastWindow >= slowWindow {
		return fmt.Errorf("fast burn window must be smaller than slow burn window")
	}

	if slowWindow > sloWindow {
		return fmt.Errorf("slow burn window must be less than or equal to SLO window")
	}

	return nil
}
