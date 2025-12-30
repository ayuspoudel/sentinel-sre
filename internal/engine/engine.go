package engine

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/ayuspoudel/sentinel-sre/internal/registry"
)

type Engine struct {
	registry       registry.Registry
	evalInterval   time.Duration
	reloadInterval time.Duration

	mu     sync.RWMutex
	guards []registry.Guard
}

func New(reg registry.Registry, evalInterval, reloadInterval time.Duration) *Engine {
	return &Engine{
		registry:       reg,
		evalInterval:   evalInterval,
		reloadInterval: reloadInterval,
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
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, g := range e.guards {
		log.Printf("evaluating guard: %s", g.Name)
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
