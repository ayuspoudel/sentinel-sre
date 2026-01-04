package engine

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/ayuspoudel/sentinel-sre/internal/action"
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
- actions
Nothing else in the system is allowed to make deployment actions.
*/
type Engine struct {
	registry registry.Registry
	metrics  *prometheus.PromClient

	evalInterval   time.Duration
	reloadInterval time.Duration

	mu      sync.RWMutex
	guards  []registry.Guard
	actions *action.Store
}

func New(reg registry.Registry, metrics *prometheus.PromClient, evalInterval, reloadInterval time.Duration) *Engine {
	return &Engine{
		registry:       reg,
		metrics:        metrics,
		evalInterval:   evalInterval,
		reloadInterval: reloadInterval,
		actions:        action.NewStore(),
	}
}

func (e *Engine) Actions() *action.Store {
	return e.actions
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
	e.mu.RLock()
	guards := e.guards
	e.mu.RUnlock()

	for _, g := range guards {

		// Phase 1: Observability
		d, ok := e.phaseObservability(ctx, g)
		if !ok {
			e.actions.Set(action.Action{
				GuardName: g.Name,
				Type:      action.Block,
				Phase:     d.Phase,
				Reason:    d.Reason,
				Timestamp: time.Now(),
			})
			log.Printf("evaluating guard: %s, %s", g.Name, d.Reason)
			continue
		}

		// Phase 2: Error budget
		d, ok = e.phaseBudget(ctx, g)
		if !ok {
			e.actions.Set(action.Action{
				GuardName: g.Name,
				Type:      action.Block,
				Phase:     d.Phase,
				Reason:    d.Reason,
				Timestamp: time.Now(),
			})
			log.Printf("evaluating guard: %s, %s", g.Name, d.Reason)
			continue
		}

		// All phases passed
		e.actions.Set(action.Action{
			GuardName: g.Name,
			Type:      action.Allow,
			Phase:     "stable",
			Reason:    "all checks passed",
			Timestamp: time.Now(),
		})

		log.Printf("evaluating guard: %s, deployment allowed", g.Name)
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
	Helpers for decisions (Deprecated because we are using actions instead of decision)
*/

// func (e *Engine) Decisions() []Decision {
// 	e.mu.RLock()
// 	defer e.mu.RUnlock()

// 	out := make([]Decision, 0, len(e.decisions))
// 	for _, d := range e.decisions {
// 		out = append(out, d)
// 	}
// 	return out
// }

// func (e *Engine) DecisionFor(guard string) (Decision, bool) {
// 	e.mu.RLock()
// 	defer e.mu.RUnlock()

// 	d, ok := e.decisions[guard]
// 	return d, ok
// }
