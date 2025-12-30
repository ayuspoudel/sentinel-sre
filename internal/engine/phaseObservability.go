package engine

import (
	"context"

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
	if value < traffic.MinRPS {
		return Decision{
			GuardName: g.Name,
			Allowed:   false,
			Phase:     "observability",
			Reason:    "insufficient traffic for reliable analysis",
		}, false

	}
	return Decision{}, false
}
