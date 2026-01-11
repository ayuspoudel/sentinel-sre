package presence

import (
	"context"

	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/install"
	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/logging"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func DetectAgentPrensence(ctx context.Context, namespace string, client *kubernetes.Clientset) (installed bool, err error) {
	log := logging.From(ctx)
	log.Info("checking agent presence", "deployment", install.AgentDeploymentName, "namespace", namespace)

	_, err = client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("namespace not found")
			return false, nil
		}
		log.Error("failed to get namespace", "error", err)
		return false, err
	}

	_, err = client.AppsV1().Deployments(namespace).Get(ctx, install.AgentDeploymentName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("deployment not found")
			return false, nil
		}
		log.Error("failed to get deployment", "error", err)
		return false, err
	}

	log.Info("agent detected")
	return true, nil
}
