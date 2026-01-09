package agent

import (
	"context"
	"log"
	"time"

	"github.com/ayuspoudel/sentinel-sre/controlplane/registryClient"
	"k8s.io/client-go/kubernetes"
)

func (c *Controller) reconcileCluster(ctx context.Context, cl *registryClient.Cluster) {
	start := time.Now()
	log.Printf("[reconcile] cluster=%s credential_ref=%s labels=%v", cl.Name, cl.CredentialRef, cl.Labels)
	status := &ClusterStatus{ClusterName: cl.Name, LastReconcileAt: &start}
	defer func() {
		duration := int(time.Since(start).Milliseconds())
		status.LastReconcileDurationMs = &duration
		err := c.store.Upsert(ctx, status)
		if err != nil {
			log.Printf("[reconcile] cluster=%s failed to upsert status (last reconcile duration upsert failed): %v", cl.Name, err)
		}
	}()

	contextName, ok := cl.Labels["context"]
	if !ok || contextName == "" {
		errMsg := "missing context label in registry stored by sentinel-k8s-cluster-registry"
		status.LastError = &errMsg
		success := false
		status.LastReconcileSuccess = &success
		log.Printf("[reconcile] cluster=%s missing context label, skipping", cl.Name)
		return
	}
	kubeconfig, err := c.loadKubeConfig(ctx, "sentinel", cl.CredentialRef)
	if err != nil {
		errMsg := "missing context label"
		status.LastError = &errMsg
		success := false
		status.LastReconcileSuccess = &success
		log.Printf("[reconcile] cluster=%s failed to load kubeconfig: %v", cl.Name, err)
		return
	}
	restCfg, err := c.buildRestConfig(kubeconfig, contextName)
	if err != nil {
		errMsg := err.Error()
		success := false
		status.LastError = &errMsg
		status.LastReconcileSuccess = &success
		log.Printf("[reconcile] cluster=%s failed to build rest config: %v", cl.Name, err)
		return
	}
	targetClient, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		errMsg := err.Error()
		success := false
		status.LastError = &errMsg
		status.LastReconcileSuccess = &success
		log.Printf("[reconcile] cluster=%s failed to create kubernetes client: %v", cl.Name, err)
		return
	}

	version, err := targetClient.Discovery().ServerVersion()
	if err != nil {
		reachable := false
		errMsg := err.Error()
		success := false
		status.Reachable = &reachable
		status.LastError = &errMsg
		status.LastReconcileSuccess = &success
		log.Printf("[reconcile] cluster=%s failed to get server version: %v", cl.Name, err)
		return
	}
	log.Printf("[reconcile] cluster=%s server version: %s", cl.Name, version.GitVersion)

}
