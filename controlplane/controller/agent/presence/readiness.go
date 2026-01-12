package presence

import (
	"context"

	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/install"
	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/logging"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func DetectAgentReadiness(ctx context.Context, namespace string, client *kubernetes.Clientset) (ready bool, err error) {
	log := logging.From(ctx)

	deploy, err := client.AppsV1().Deployments(namespace).Get(ctx, install.AgentDeploymentName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("agent deployment not found for readiness check")
			return false, nil
		}
		log.Error("failed to get agent deployment", "error", err)
		return false, err
	}

	if deploy.Spec.Replicas == nil {
		log.Warn("agent deployment replicas not set")
		return false, nil
	}

	if deploy.Status.ObservedGeneration < deploy.Generation {
		log.Info("deployment not yet observed", "observed", deploy.Status.ObservedGeneration, "generation", deploy.Generation)
		return false, nil
	}

	if deploy.Status.ReadyReplicas < *deploy.Spec.Replicas {
		log.Info(
			"agent deployment not ready",
			"ready", deploy.Status.ReadyReplicas,
			"desired", *deploy.Spec.Replicas,
		)
		return false, nil
	}

	log.Info("agent deployment is ready")
	return true, nil
}
