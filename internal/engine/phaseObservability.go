package engine

import (
	"context"
	"log"

	"github.com/ayuspoudel/sentinel-sre/internal/registry"
)

func (e *Engine) phaseObservability(ctx context.Context, g registry.Guard) (Decision, bool) {
	traffic := g.Manifest.Signals.Traffic
	value, err := e.metrics.Query(ctx, traffic.Query)
	if err != nil {
		return Decision{
			GuardName: g.Name,
			Allowed:   false,
			Phase:     "observability",
			Reason:    "failed to query prometheus: " + err.Error(),
		}, false
	}
	log.Printf(
		"guard=%s traffic_rps=%f min_rps=%f",
		g.Name,
		value,
		g.Manifest.Signals.Traffic.MinRPS,
	)
	if value < traffic.MinRPS {
		return Decision{
			GuardName: g.Name,
			Allowed:   false,
			Phase:     "observability",
			Reason:    "insufficient traffic for reliable analysis",
		}, false

	}
	return Decision{
		GuardName: g.Name,
		Allowed:   true,
		Phase:     "observability",
		Reason:    "sufficient traffic observed",
	}, true
}
