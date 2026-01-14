package controller

import (
	"context"
	"reflect"
	"time"

	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/adoption"
	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/install"
	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/kube"
	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/logging"
	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/presence"
	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/status"
	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/values"
	"github.com/ayuspoudel/sentinel-sre/controlplane/registryClient"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

func (c *Controller) reconcileCluster(ctx context.Context, cl *registryClient.Cluster) {
	log := logging.From(ctx)
	start := time.Now()

	st := &status.ClusterStatus{
		ClusterName:     cl.Name,
		LastReconcileAt: &start,
	}

	defer func() {
		duration := int(time.Since(start).Milliseconds())
		log.Info("reconcile completed", "duration_ms", duration, "success", st.LastReconcileSuccess)
	}()

	log.Info("reconcile started", "labels", cl.Labels)

	defer func() {
		duration := int(time.Since(start).Milliseconds())
		st.LastReconcileDurationMs = &duration

		existing, err := c.store.Get(ctx, cl.Name)
		if err != nil {
			log.Warn("failed to load existing cluster status, proceeding with new", "error", err)
		}

		merged := status.Merge(existing, st)

		if err := c.store.Upsert(ctx, merged); err != nil {
			log.Error("failed to upsert merged cluster status", "error", err)
		}
		if hasMeaningfulChange(existing, merged) {
			if err := c.publisher.PublishClusterStatus(ctx, merged); err != nil {
				log.Warn(
					"failed to publish cluster status event",
					"cluster", merged.ClusterName,
					"error", err,
				)
			}
		}
	}()

	contextName, ok := cl.Labels["context"]
	if !ok || contextName == "" {
		errMsg := "missing context label in registry stored by sentinel-k8s-cluster-registry"
		success := false
		st.LastError = &errMsg
		st.LastReconcileSuccess = &success
		log.Warn("missing context label, skipping reconcile", "required_label", "context")
		return
	}

	kubeconfig, err := kube.LoadKubeConfig(ctx, c.kubeClient, "sentinel", cl.CredentialRef)
	if err != nil {
		errMsg := err.Error()
		success := false
		st.LastError = &errMsg
		st.LastReconcileSuccess = &success
		log.Error("failed to load kubeconfig", "namespace", "sentinel", "secret", cl.CredentialRef, "error", err)
		return
	}

	restCfg, err := kube.BuildRestConfig(kubeconfig, contextName)
	if err != nil {
		errMsg := err.Error()
		success := false
		st.LastError = &errMsg
		st.LastReconcileSuccess = &success
		log.Error("failed to build rest config", "context", contextName, "error", err)
		return
	}

	targetClient, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		errMsg := err.Error()
		success := false
		st.LastError = &errMsg
		st.LastReconcileSuccess = &success
		log.Error("failed to create kubernetes client", "error", err)
		return
	}

	version, err := targetClient.Discovery().ServerVersion()
	if err != nil {
		reachable := false
		success := false
		errMsg := err.Error()
		st.Reachable = &reachable
		st.LastError = &errMsg
		st.LastReconcileSuccess = &success
		log.Error("failed to connect to api server", "error", err)
		return
	}

	reachable := true
	authValid := true
	success := true
	now := time.Now()

	st.Reachable = &reachable
	st.AuthValid = &authValid
	st.LastReconcileSuccess = &success
	st.APIServerVersion = &version.GitVersion
	st.LastSuccessfulConnection = &now

	log.Info("connected to api server", "version", version.GitVersion)

	installed, err := presence.DetectAgentPrensence(ctx, install.AgentNamespace, targetClient)
	if err != nil {
		errMsg := err.Error()
		success := false
		st.LastError = &errMsg
		st.LastReconcileSuccess = &success
		log.Error("failed to detect agent presence", "error", err)
		return
	}

	agentID := cl.Name
	agentVersion := install.AgentImageTag

	if !installed {
		log.Info("agent not present, installing")

		dynClient, err := dynamic.NewForConfig(restCfg)
		if err != nil {
			errMsg := err.Error()
			success := false
			st.LastError = &errMsg
			st.LastReconcileSuccess = &success
			log.Error("failed to create dynamic client", "error", err)
			return
		}

		agentValues := values.BuildAgentValues(
			c.controlPlaneUrl,
			c.controlPlaneUrl,
			cl.Name,
			agentID,
			agentVersion,
			install.AgentImageRepo,
			install.AgentImageTag,
		)

		manifest, err := adoption.RenderManifests(
			ctx,
			kubeconfig,
			contextName,
			install.AgentNamespace,
			install.AgentHelmRepo+"/"+install.AgentDeploymentName,
			agentValues,
		)
		if err != nil {
			errMsg := err.Error()
			success := false
			st.LastError = &errMsg
			st.LastReconcileSuccess = &success
			log.Error("failed to render helm manifests", "error", err)
			return
		}

		objects, err := adoption.ParseManifests([]byte(manifest))
		if err != nil {
			errMsg := err.Error()
			success := false
			st.LastError = &errMsg
			st.LastReconcileSuccess = &success
			log.Error("failed to parse helm manifests", "error", err)
			return
		}

		for _, obj := range objects {
			if err := adoption.Adopt(ctx, dynClient, obj, adoption.Ownership{
				ReleaseName:      install.AgentReleaseName,
				ReleaseNamespace: install.AgentNamespace,
			}); err != nil {
				errMsg := err.Error()
				success := false
				st.LastError = &errMsg
				st.LastReconcileSuccess = &success
				log.Error("failed to adopt resource", "error", err)
				return
			}
		}

		if err := c.installer.Install(ctx, &install.InstallConfig{
			KubeConfig:  kubeconfig,
			ContextName: contextName,
			Values:      agentValues,
		}); err != nil {
			errMsg := err.Error()
			success := false
			st.LastError = &errMsg
			st.LastReconcileSuccess = &success
			log.Error("failed to install agent", "error", err)
			return
		}
	}

	if installed {
		st.AgentInstalled = ptr(true)
		st.AgentNamespace = ptr(install.AgentNamespace)
	}

	ready, err := presence.DetectAgentReadiness(ctx, install.AgentNamespace, targetClient)
	if err != nil {
		errMsg := err.Error()
		success := false
		st.LastError = &errMsg
		st.LastReconcileSuccess = &success
		log.Error("failed to detect agent readiness", "error", err)
		return
	}

	st.AgentHealthy = ptr(ready)
	st.LastReconcileSuccess = ptr(true)

	log.Info("reconcile finished successfully", "agent_ready", ready)
}

func ptr[T any](v T) *T {
	return &v
}

/*
Author: @ayuspoudel
This function evaluates if there are any meaningful change in the cluster status
If yes, then only it will allow any DB insert or Event Publication. It is useful
to avoid noise, unnecessary db writes and event publish.
*/
func hasMeaningfulChange(old, new *status.ClusterStatus) bool {
	if old == nil {
		return true
	}

	oldSignal := map[string]any{
		"reachable":       old.Reachable,
		"auth_valid":      old.AuthValid,
		"agent_installed": old.AgentInstalled,
		"agent_healthy":   old.AgentHealthy,
		"agent_version":   old.AgentVersion,
		"agent_namespace": old.AgentNamespace,
	}

	newSignal := map[string]any{
		"reachable":       new.Reachable,
		"auth_valid":      new.AuthValid,
		"agent_installed": new.AgentInstalled,
		"agent_healthy":   new.AgentHealthy,
		"agent_version":   new.AgentVersion,
		"agent_namespace": new.AgentNamespace,
	}

	return !reflect.DeepEqual(oldSignal, newSignal)
}
