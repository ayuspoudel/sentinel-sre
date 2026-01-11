package controller

import (
	"context"
	"log"
	"time"

	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/adoption"
	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/install"
	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/kube"
	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/presence"
	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/status"
	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/values"
	"github.com/ayuspoudel/sentinel-sre/controlplane/registryClient"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

func (c *Controller) reconcileCluster(ctx context.Context, cl *registryClient.Cluster) {
	start := time.Now()
	log.Printf("[reconcile] cluster=%s credential_ref=%s labels=%v", cl.Name, cl.CredentialRef, cl.Labels)
	status := &status.ClusterStatus{ClusterName: cl.Name, LastReconcileAt: &start}
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
	kubeconfig, err := kube.LoadKubeConfig(ctx, c.kubeClient, "sentinel", cl.CredentialRef)
	if err != nil {
		errMsg := err.Error()
		status.LastError = &errMsg
		success := false
		status.LastReconcileSuccess = &success
		log.Printf("[reconcile] cluster=%s failed to load kubeconfig: %v", cl.Name, err)
		return
	}
	restCfg, err := kube.BuildRestConfig(kubeconfig, contextName)
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
		success := false
		errMsg := err.Error()

		status.Reachable = &reachable
		status.LastError = &errMsg
		status.LastReconcileSuccess = &success
		return
	}

	reachable := true
	authValid := true
	success := true
	now := time.Now()

	status.Reachable = &reachable
	status.AuthValid = &authValid
	status.LastReconcileSuccess = &success
	status.APIServerVersion = &version.GitVersion
	status.LastSuccessfulConnection = &now

	installed, namespace, err := presence.DetectAgentPrensence(ctx, targetClient)
	if err != nil {
		errMsg := err.Error()
		success := false
		status.LastError = &errMsg
		status.LastReconcileSuccess = &success
		log.Printf("[reconcile] cluster=%s failed to detect agent presence: %v", cl.Name, err)
		return
	}

	if !installed {
		log.Printf("[reconcile] cluster=%s agent not present, installing", cl.Name)
		dynClient, err := dynamic.NewForConfig(restCfg)
		if err != nil {
			errMsg := err.Error()
			success := false
			status.LastError = &errMsg
			status.LastReconcileSuccess = &success
			log.Printf("[reconcile] cluster=%s failed to create dynamic client: %v", cl.Name, err)
			return
		}

		manifest, err := adoption.RenderManifests(
			ctx,
			kubeconfig,
			contextName,
			install.AgentNamespace,
			install.AgentHelmRepo+"/"+install.AgentDeploymentName,
			values.BuildAgentValues(c.controlPlaneUrl, cl.Name),
		)
		if err != nil {
			errMsg := err.Error()
			success := false
			status.LastError = &errMsg
			status.LastReconcileSuccess = &success
			log.Printf("[reconcile] cluster=%s failed to render helm manifests: %v", cl.Name, err)
			return
		}

		objects, err := adoption.ParseManifests([]byte(manifest))

		if err != nil {
			errMsg := err.Error()
			success := false
			status.LastError = &errMsg
			status.LastReconcileSuccess = &success
			log.Printf("[reconcile] cluster=%s failed to parse helm manifests: %v", cl.Name, err)
			return
		}

		for _, obj := range objects {
			err := adoption.Adopt(ctx, dynClient, obj, adoption.Ownership{
				ReleaseName:      install.AgentReleaseName,
				ReleaseNamespace: install.AgentNamespace,
			})
			if err != nil {
				errMsg := err.Error()
				success := false
				status.LastError = &errMsg
				status.LastReconcileSuccess = &success
				log.Printf("[reconcile] cluster=%s failed to adopt resource: %v", cl.Name, err)
				return
			}
		}

		values := values.BuildAgentValues(c.controlPlaneUrl, cl.Name)
		err = c.installer.Install(ctx, &install.InstallConfig{
			KubeConfig:  kubeconfig,
			ContextName: contextName,
			Values:      values,
		})
		if err != nil {
			errMsg := err.Error()
			success := false
			status.LastError = &errMsg
			status.LastReconcileSuccess = &success
			log.Printf("[reconcile] cluster=%s failed to install agent: %v", cl.Name, err)
			return
		}
	}
	success = true
	installed = true
	status.AgentInstalled = &installed
	status.AgentNamespace = namespace
	status.LastReconcileSuccess = &success
	log.Printf("[reconcile] cluster=%s reconcile completed successfully", cl.Name)

}
