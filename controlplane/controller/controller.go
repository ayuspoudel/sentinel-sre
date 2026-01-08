package controller

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/ayuspoudel/sentinel-sre/internal/prometheus"
)

type Decision struct {
	BlockDeploy bool
	Reason      string
	ErrorRate   float64
}

type Controller struct {
	prom         *prometheus.PromClient
	maxErrorRate float64

	mu     sync.RWMutex
	latest Decision
}

func New(prom *prometheus.PromClient, maxErrorRate float64) *Controller {
	return &Controller{
		prom:         prom,
		maxErrorRate: maxErrorRate,
		latest: Decision{
			BlockDeploy: false,
			Reason:      "starting up, no data yet",
			ErrorRate:   -1,
		},
	}
}

func (c *Controller) Start(ctx context.Context) {

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.evaluateOnce(ctx)
		case <-ctx.Done():
			log.Println("controller stopped")
			return
		}
	}
}

/*
	@ayuspoudel
	This is one of the most important function in this util. It represents one
	decision making cycle of the controller. In one run it:
		- asks prom how much traffic is system handling
		- asks prom how many of those req are failing
		- interprets those to understand sys health
		- decides whether deployments should be blocked or allowed
		- stores that decision safely inside container
	In sum, everytime this function runs, the controller looks at current state of
	the system and updates its opinion about whether its safe to deploy.
*/

func (c *Controller) evaluateOnce(ctx context.Context) {
	total, err := c.prom.Query(ctx, `sum(rate(http_requests_total[1m]))`)
	if err != nil {
		log.Printf("total query failed: %v", err)
		return
	}
	errors, err := c.prom.Query(ctx, `sum(rate(http_requests_total{status=~"5.."}[1m]))`)
	if err != nil {
		log.Printf("total 5xx query failed: %v", err)
		return
	}
	if total == 0 {
		c.setDecision(Decision{BlockDeploy: false, Reason: "no traffic", ErrorRate: 0})
		return
	}
	errorRate := errors / total

	if errorRate > c.maxErrorRate {
		c.setDecision(Decision{BlockDeploy: true, Reason: "high error rate", ErrorRate: errorRate})
	} else {
		c.setDecision(Decision{BlockDeploy: false, Reason: "error rate within limits", ErrorRate: errorRate})
	}
}

/*
	Utility functions
*/

func (c *Controller) setDecision(d Decision) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.latest = d
}

func (c *Controller) LatestDecision() Decision {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.latest
}
