package engine

import (
	"context"
	"log"
	"sync"
	"time"

	prometheus "github.com/ayuspoudel/sentinel-sre/internal/prometheus"
	"github.com/ayuspoudel/sentinel-sre/internal/registry"
)

/*
@ayuspoudel
Engine is the core of Sentinel.
It owns:
- registry (source of guards)
- prometheus client (metrics source)
- evaluation loop
- decisions
Nothing else in the system is allowed to make deployment decisions.
*/
type Engine struct {
	registry registry.Registry
	metrics  *prometheus.PromClient

	evalInterval   time.Duration
	reloadInterval time.Duration

	mu        sync.RWMutex
	guards    []registry.Guard
	decisions map[string]Decision
}

func New(reg registry.Registry, metrics *prometheus.PromClient, evalInterval, reloadInterval time.Duration) *Engine {
	return &Engine{
		registry:       reg,
		metrics:        metrics,
		evalInterval:   evalInterval,
		reloadInterval: reloadInterval,
		decisions:      make(map[string]Decision),
	}
}

func (e *Engine) Start(ctx context.Context) error {
	err := e.registry.Load(ctx)
	if err != nil {
		return err
	}

	e.guards = e.registry.Guards()
	evalTicker := time.NewTicker(e.evalInterval)
	reloadTicker := time.NewTicker(e.reloadInterval)
	defer evalTicker.Stop()
	defer reloadTicker.Stop()
	log.Printf("engine started with %d guards", len(e.guards))
	for {
		select {
		case <-evalTicker.C:
			e.evaluateOnce(ctx)
		case <-reloadTicker.C:
			e.tryReload(ctx)
		case <-ctx.Done():
			log.Println("engine stopped")
			return nil
		}
	}
}

func (e *Engine) evaluateOnce(ctx context.Context) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, g := range e.guards {

		// Phase 1: Observability
		d, ok := e.phaseObservability(ctx, g)
		d.Timestamp = time.Now()
		e.decisions[g.Name] = d
		log.Printf("evaluating guard: %s, %s", g.Name, d.Reason)

		if !ok {
			continue
		}

		// Phase 2: Error budget
		d, ok = e.phaseBudget(ctx, g)
		d.Timestamp = time.Now()
		e.decisions[g.Name] = d
		log.Printf("evaluating guard: %s, %s", g.Name, d.Reason)

		if !ok {
			continue
		}

		// All phases passed
		e.decisions[g.Name] = Decision{
			GuardName: g.Name,
			Allowed:   true,
			Phase:     "final",
			Reason:    "all checks passed",
			Timestamp: time.Now(),
		}
	}
}

func (e *Engine) tryReload(ctx context.Context) {
	log.Println("attempting registry reload")
	err := e.reload(ctx)
	if err != nil {
		log.Printf("registry reload failed: %v (keeping previous state)", err)
		return
	}
	log.Printf("registry reload succeeded, %d guards active", len(e.guards))

}

func (e *Engine) reload(ctx context.Context) error {
	err := e.registry.Load(ctx)
	if err != nil {
		return err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.guards = e.registry.Guards()
	return nil

}

/*
	Helpers for decisions
*/

func (e *Engine) Decisions() []Decision {
	e.mu.RLock()
	defer e.mu.RUnlock()

	out := make([]Decision, 0, len(e.decisions))
	for _, d := range e.decisions {
		out = append(out, d)
	}
	return out
}

func (e *Engine) DecisionFor(guard string) (Decision, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	d, ok := e.decisions[guard]
	return d, ok
}
