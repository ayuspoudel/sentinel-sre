package controller

import (
	"context"
	"log"
	"time"

	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/install"
	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/status"
	"github.com/ayuspoudel/sentinel-sre/controlplane/registryClient"
	"k8s.io/client-go/kubernetes"
)

type Controller struct {
	registry        *registryClient.Client
	kubeClient      *kubernetes.Clientset
	store           status.Store
	interval        time.Duration
	controlPlaneUrl string
	installer       install.Installer
}

func NewController(registry *registryClient.Client, kubeClient *kubernetes.Clientset, store status.Store, interval time.Duration, controlPlaneUrl string, installer install.Installer) *Controller {
	return &Controller{registry: registry, kubeClient: kubeClient, store: store, interval: interval, controlPlaneUrl: controlPlaneUrl, installer: installer}
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
