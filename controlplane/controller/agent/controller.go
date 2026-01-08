package agent

import (
	"context"
	"log"
	"time"

	"github.com/ayuspoudel/sentinel-sre/controlplane/registryClient"
)

type Controller struct {
	registry *registryClient.Client
	interval time.Duration
}

func NewController(registry *registryClient.Client, interval time.Duration) *Controller {
	return &Controller{registry: registry, interval: interval}
}

func (c *Controller) Run(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("sentinel agent controller stopping")
			return
		case <-ticker.C:
			c.reconcileOnce(ctx)
		}
	}
}

func (c *Controller) reconcileOnce(ctx context.Context) {
	clusters, err := c.registry.ListClusters(ctx)
	if err != nil {
		log.Printf("error listing clusters: %v", err)
		return
	}
	for _, cluster := range clusters {
		log.Printf("reconciling cluster: %s", cluster.Name)
		c.reconcileCluster(ctx, cluster)
	}
}
