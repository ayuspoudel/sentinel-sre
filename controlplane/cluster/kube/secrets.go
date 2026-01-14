package kube

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func EnsureKubeconfigSecret(ctx context.Context, client *kubernetes.Clientset, namespace, name string, kubeconfig []byte) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"kubeconfig": kubeconfig,
		},
	}
	_, err := client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			_, err = client.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			}, metav1.CreateOptions{})
		}
		if err != nil {
			return fmt.Errorf("failed to ensure namespace for kubeconfig secret: %w", err)
		}
	}
	_, err = client.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, updateErr := client.CoreV1().Secrets(namespace).Update(ctx, secret, metav1.UpdateOptions{})
			if updateErr != nil {
				return fmt.Errorf("failed to update existing kubeconfig: %w", updateErr)
			}
		} else {
			return fmt.Errorf("failed to ensure kubeconfig secret: %w", err)
		}
	}
	return nil

}
