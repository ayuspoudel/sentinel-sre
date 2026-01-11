package controller

import (
	"context"
	"time"

	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/adoption"
	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/install"
	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/kube"
	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/logging"
	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/presence"
	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/status"
	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/values"
	"github.com/ayuspoudel/sentinel-sre/controlplane/registryClient"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

func (c *Controller) reconcileCluster(ctx context.Context, cl *registryClient.Cluster) {
	log := logging.From(ctx)
	start := time.Now()
	status := &status.ClusterStatus{ClusterName: cl.Name, LastReconcileAt: &start}
	defer func() {
		duration := int(time.Since(start).Milliseconds())
		log.Info("reconcile completed", "duration_ms", duration, "success", status.LastReconcileSuccess)
	}()

	log.Info("reconcile started", "labels", cl.Labels)
	defer func() {
		duration := int(time.Since(start).Milliseconds())
		status.LastReconcileDurationMs = &duration
		err := c.store.Upsert(ctx, status)
		if err != nil {
			log.Error("failed to upsert cluster status", "error", err)
		}
	}()

	contextName, ok := cl.Labels["context"]
	if !ok || contextName == "" {
		errMsg := "missing context label in registry stored by sentinel-k8s-cluster-registry"
		status.LastError = &errMsg
		success := false
		status.LastReconcileSuccess = &success
		log.Warn("missing context label, skipping reconcile", "required_label", "context")
		return
	}

	kubeconfig, err := kube.LoadKubeConfig(ctx, c.kubeClient, "sentinel", cl.CredentialRef)
	if err != nil {
		errMsg := err.Error()
		status.LastError = &errMsg
		success := false
		status.LastReconcileSuccess = &success
		log.Error("failed to load kubeconfig", "namespace", "sentinel", "secret", cl.CredentialRef, "error", err)
		return
	}

	restCfg, err := kube.BuildRestConfig(kubeconfig, contextName)
	if err != nil {
		errMsg := err.Error()
		success := false
		status.LastError = &errMsg
		status.LastReconcileSuccess = &success
		log.Error("failed to build rest config", "context", contextName, "error", err)
		return
	}

	targetClient, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		errMsg := err.Error()
		success := false
		status.LastError = &errMsg
		status.LastReconcileSuccess = &success
		log.Error("failed to create kubernetes client", "error", err)
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
		log.Error("failed to connect to api server", "error", err)
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

	log.Info("connected to api server", "version", version.GitVersion)

	installed, err := presence.DetectAgentPrensence(ctx, install.AgentNamespace, targetClient)
	if err != nil {
		errMsg := err.Error()
		success := false
		status.LastError = &errMsg
		status.LastReconcileSuccess = &success
		log.Error("failed to detect agent presence", "error", err)
		return
	}

	if !installed {
		log.Info("agent not present, installing")

		dynClient, err := dynamic.NewForConfig(restCfg)
		if err != nil {
			errMsg := err.Error()
			success := false
			status.LastError = &errMsg
			status.LastReconcileSuccess = &success
			log.Error("failed to create dynamic client", "error", err)
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
			log.Error("failed to render helm manifests", "error", err)
			return
		}

		objects, err := adoption.ParseManifests([]byte(manifest))
		if err != nil {
			errMsg := err.Error()
			success := false
			status.LastError = &errMsg
			status.LastReconcileSuccess = &success
			log.Error("failed to parse helm manifests", "error", err)
			return
		}

		for _, obj := range objects {
			err := adoption.Adopt(ctx, dynClient, obj, adoption.Ownership{
				ReleaseName: install.AgentReleaseName, ReleaseNamespace: install.AgentNamespace,
			})
			if err != nil {
				errMsg := err.Error()
				success := false
				status.LastError = &errMsg
				status.LastReconcileSuccess = &success
				log.Error("failed to adopt resource", "error", err)
				return
			}
		}

		values := values.BuildAgentValues(c.controlPlaneUrl, cl.Name)
		err = c.installer.Install(ctx, &install.InstallConfig{
			KubeConfig: kubeconfig, ContextName: contextName, Values: values,
		})
		if err != nil {
			errMsg := err.Error()
			success := false
			status.LastError = &errMsg
			status.LastReconcileSuccess = &success
			log.Error("failed to install agent", "error", err)
			return
		}
	}

	if installed {
		status.AgentInstalled = ptr(true)
		status.AgentNamespace = ptr(install.AgentNamespace)
	}

	status.LastReconcileSuccess = ptr(true)
	log.Info("reconcile finished successfully")
}

func ptr[T any](v T) *T {
	return &v
}
