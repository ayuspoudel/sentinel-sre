package agent

import (
	"context"
	"log"
	"time"

	"github.com/ayuspoudel/sentinel-sre/controlplane/registryClient"
	"k8s.io/client-go/kubernetes"
)

type Controller struct {
	registry   *registryClient.Client
	kubeClient *kubernetes.Clientset
	store      Store
	interval   time.Duration
}

func NewController(registry *registryClient.Client, kubeClient *kubernetes.Clientset, store Store, interval time.Duration) *Controller {
	return &Controller{registry: registry, kubeClient: kubeClient, store: store, interval: interval}
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
