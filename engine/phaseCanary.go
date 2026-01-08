package engine

import (
	"context"
	"log"
	"time"

	"github.com/ayuspoudel/sentinel-sre/internal/registry"
)

func (e *Engine) phaseCanary(ctx context.Context, g registry.Guard) (Decision, bool) {
	canary := g.Manifest.Canary
	if canary == nil || !canary.Enabled {
		return Decision{
			GuardName: g.Name,
			Allowed:   true,
			Phase:     "canary",
			Reason:    "canary analysis disabled",
		}, true
	}

	errors := g.Manifest.Signals.Errors
	slo := g.Manifest.Signals.SLO
	budget := g.Manifest.Policy.Budget
	allowedErrRatio := 1 - (slo.Objective / 100)

	canaryQuery := applyScope(errors.Query, canary.Scope.Label, canary.Scope.CanaryValue)
	baselineQuery := applyScope(errors.Query, canary.Scope.Label, canary.Scope.BaselineValue)

	canaryBurnQuery := buildBurnQuery(canaryQuery, budget.FastBurn.Window, allowedErrRatio)
	baselineBurnQuery := buildBurnQuery(baselineQuery, budget.FastBurn.Window, allowedErrRatio)

	canaryBurn, err := e.metrics.Query(ctx, canaryBurnQuery)

	if err != nil {
		return Decision{
			GuardName: g.Name,
			Allowed:   false,
			Phase:     "canary",
			Reason:    "failed to query prometheus for canary burn: " + err.Error(),
		}, false

	}
	baselineBurn, err := e.metrics.Query(ctx, baselineBurnQuery)
	if err != nil {
		return Decision{
			GuardName: g.Name,
			Allowed:   false,
			Phase:     "canary",
			Reason:    "failed to query prometheus for baseline burn: " + err.Error(),
		}, false
	}

	log.Printf("guard=%s canary_burn=%f baseline_burn=%f", g.Name, canaryBurn, baselineBurn)

	if canaryBurn > baselineBurn*1.2 {
		return Decision{
			GuardName: g.Name,
			Allowed:   false,
			Phase:     "canary",
			Reason:    "canary burn exceeds baseline",
			Timestamp: time.Now(),
		}, false
	}

	return Decision{
		GuardName: g.Name,
		Allowed:   true,
		Phase:     "canary",
		Reason:    "canary within acceptable deviation",
		Timestamp: time.Now(),
	}, true
}
