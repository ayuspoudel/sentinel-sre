package kube

import (
	"context"
	"fmt"

	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/logging"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func LoadKubeConfig(ctx context.Context, client *kubernetes.Clientset, namespace, secretName string) (kubeconfig []byte, err error) {
	log := logging.From(ctx)
	log.Info("loading kubeconfig secret", "namespace", namespace, "secret", secretName)

	secret, err := client.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Warn("kubeconfig secret not found")
			return nil, fmt.Errorf("kubeconfig secret %s/%s not found", namespace, secretName)
		}
		log.Error("failed to get kubeconfig secret", "error", err)
		return nil, fmt.Errorf("failed to get kubeconfig secret %s/%s: %w", namespace, secretName, err)
	}

	kubeconfig, ok := secret.Data["kubeconfig"]
	if !ok {
		log.Error("kubeconfig key missing in secret")
		return nil, fmt.Errorf("secret %s missing kubeconfig", secretName)
	}

	log.Info("kubeconfig loaded successfully")
	return kubeconfig, nil
}

/*
@ayuspoudel
Kubernetes clients in Go do not talk to k8s directly using kubeconfig file
They talk using *rest.Config object which has:
  - API Server URL
  - TLS Certs/CA
  - Auth Info (token, client cert etc)
  - Timeout, QPS, retries

This function converts raw kubeconfig file + choosen contextName in kubeconfig
file into a rest client object which our go server can use.
*/
func BuildRestConfig(kubeconfig []byte, contextName string) (*rest.Config, error) {
	log := logging.From(context.Background())
	log.Info("building rest config from kubeconfig", "context", contextName)

	cfg, err := clientcmd.Load(kubeconfig)
	if err != nil {
		log.Error("failed to load kubeconfig", "error", err)
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	_, ok := cfg.Contexts[contextName]
	if !ok {
		log.Error("context not found in kubeconfig", "context", contextName)
		return nil, fmt.Errorf("context %s not found in kubeconfig", contextName)
	}

	overrides := &clientcmd.ConfigOverrides{CurrentContext: contextName}
	clientCfg := clientcmd.NewDefaultClientConfig(*cfg, overrides)

	restCfg, err := clientCfg.ClientConfig()
	if err != nil {
		log.Error("failed to build rest config", "error", err)
		return nil, err
	}

	log.Info("rest config built successfully", "context", contextName)
	return restCfg, nil
}
