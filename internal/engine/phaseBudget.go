package engine

import (
	"context"
	"log"

	"github.com/ayuspoudel/sentinel-sre/internal/registry"
)

func (e *Engine) phaseBudget(ctx context.Context, g registry.Guard) (Decision, bool) {
	budget := g.Manifest.Policy.Budget
	slo := g.Manifest.Signals.SLO
	errors := g.Manifest.Signals.Errors
	allowedErrRatio := 1 - (slo.Objective / 100)

	fastQuery := buildBurnQuery(errors.Query, budget.FastBurn.Window, allowedErrRatio)
	fastBurn, err := e.metrics.Query(ctx, fastQuery)
	if err != nil {
		return Decision{
			GuardName: g.Name,
			Allowed:   false,
			Phase:     "budget",
			Reason:    "failed to query prometheus for fast burn: " + err.Error(),
		}, false
	}

	slowQuery := buildBurnQuery(errors.Query, budget.SlowBurn.Window, allowedErrRatio)
	slowBurn, err := e.metrics.Query(ctx, slowQuery)
	if err != nil {
		return Decision{
			GuardName: g.Name,
			Allowed:   false,
			Phase:     "budget",
			Reason:    "failed to query prometheus for slow burn: " + err.Error(),
		}, false
	}

	log.Printf("guard=%s fast_burn=%f threshold=%f", g.Name, fastBurn, budget.FastBurn.Threshold)
	log.Printf("guard=%s slow_burn=%f threshold=%f", g.Name, slowBurn, budget.SlowBurn.Threshold)
	if fastBurn >= budget.FastBurn.Threshold {
		return Decision{
			GuardName: g.Name,
			Allowed:   false,
			Phase:     "budget",
			Reason:    "fast burn threshold exceeded",
		}, false
	}

	if slowBurn >= budget.SlowBurn.Threshold {
		return Decision{
			GuardName: g.Name,
			Allowed:   false,
			Phase:     "budget",
			Reason:    "slow burn threshold exceeded",
		}, false

	}

	return Decision{
		GuardName: g.Name,
		Allowed:   true,
		Phase:     "budget",
		Reason:    "error budget consumption within limits",
	}, true

}
