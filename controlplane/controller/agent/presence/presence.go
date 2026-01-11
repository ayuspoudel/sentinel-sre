package presence

import (
	"context"

	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/install"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func DetectAgentPrensence(ctx context.Context, client *kubernetes.Clientset) (installed bool, namespace *string, err error) {
	_, err = client.CoreV1().Namespaces().Get(ctx, install.AgentNamespace, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil, nil
		}
		return false, nil, err
	}
	_, err = client.AppsV1().Deployments(install.AgentNamespace).Get(ctx, install.AgentDeploymentName, metav1.GetOptions{})

	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil, nil
		}
		return false, nil, err
	}
	ns := install.AgentNamespace
	return true, &ns, nil
}
