package agent

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func detectAgentPrensence(ctx context.Context, client *kubernetes.Clientset) (installed bool, namespace *string, err error) {
	_, err = client.CoreV1().Namespaces().Get(ctx, AgentNamespace, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil, nil
		}
		return false, nil, err
	}
	_, err = client.AppsV1().Deployments(AgentNamespace).Get(ctx, AgentDeploymentName, metav1.GetOptions{})

	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil, nil
		}
		return false, nil, err
	}

	ns := AgentNamespace
	return true, &ns, nil

}
