package controller

import (
	"context"
	"time"

	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/events/clusterStatusEvent"
	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/install"
	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/logging"
	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/models/clusterStatusModel"
	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/store/clusterRegistry"
	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/store/clusterStatus"
	"k8s.io/client-go/kubernetes"
)

type Controller struct {
	kubeClient      *kubernetes.Clientset
	store           clusterStatus.Store
	clusterStatus   clusterStatusModel.ClusterStatus
	clusterRegistry clusterRegistry.Store
	interval        time.Duration
	controlPlaneUrl string
	installer       install.Installer
	publisher       clusterStatusEvent.ClusterStatusPublisher
}

func NewController(kubeClient *kubernetes.Clientset, store clusterStatus.Store, clusterRegistry clusterRegistry.Store, interval time.Duration, controlPlaneUrl string, installer install.Installer,
	publisher clusterStatusEvent.ClusterStatusPublisher) *Controller {
	return &Controller{kubeClient: kubeClient, store: store, clusterRegistry: clusterRegistry, interval: interval, controlPlaneUrl: controlPlaneUrl, installer: installer,
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
	clusters, err := c.clusterRegistry.List(ctx)
	if err != nil {
		log.Error("error listing clusters", "error", err.Error())

		return
	}
	for _, cluster := range clusters {
		clusterCtx := logging.With(ctx, "cluster", cluster.ClusterName, "credential_ref", cluster.CredentialRef)
		c.reconcileCluster(clusterCtx, cluster)
	}
}
