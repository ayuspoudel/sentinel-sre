package engine

import (
	"context"
	"log"
	"time"

	"github.com/ayuspoudel/sentinel-sre/internal/registry"
)

type Engine struct {
	registry registry.Registry
	interval time.Duration
	guards   []registry.Guard
}

func New(reg registry.Registry, interval time.Duration) *Engine {
	return &Engine{
		registry: reg,
		interval: interval,
	}
}

func (e *Engine) Start(ctx context.Context) error {
	err := e.registry.Load(ctx)
	if err != nil {
		return err
	}
	e.guards = e.registry.Guards()
	log.Printf("engine started with %d guards", len(e.guards))
	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			e.runOnce(ctx)
		case <-ctx.Done():
			log.Println("engine stopped")
			return nil
		}
	}
}

func (e *Engine) runOnce(ctx context.Context) {
	for _, g := range e.guards {
		log.Printf("evaluating guard: %s", g.Name)
	}
}
