package controller

import (
	"context"
	"time"

	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/events"
	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/install"
	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/logging"
	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/status"
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
	publisher       events.ClusterStatusPublisher
}

func NewController(registry *registryClient.Client, kubeClient *kubernetes.Clientset, store status.Store, interval time.Duration, controlPlaneUrl string, installer install.Installer,
	publisher events.ClusterStatusPublisher) *Controller {
	return &Controller{registry: registry, kubeClient: kubeClient, store: store, interval: interval, controlPlaneUrl: controlPlaneUrl, installer: installer,
		publisher: publisher}
}

func (c *Controller) Run(ctx context.Context) {
	ctx = logging.With(ctx, "component", "agent-controller")
	log := logging.From(ctx)
	log.Info("controller started", "interval", c.interval.String())
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Info("controller stopping")
			return
		case <-ticker.C:
			c.reconcileOnce(ctx)
		}
	}
}

func (c *Controller) reconcileOnce(ctx context.Context) {
	log := logging.From(ctx)
	clusters, err := c.registry.ListClusters(ctx)
	if err != nil {
		log.Error("error listing clusters", "error", err.Error())

		return
	}
	for _, cluster := range clusters {
		clusterCtx := logging.With(ctx, "cluster", cluster.Name, "credential_ref", cluster.CredentialRef)
		c.reconcileCluster(clusterCtx, cluster)
	}
}
