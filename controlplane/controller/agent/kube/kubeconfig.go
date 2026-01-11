package kube

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func LoadKubeConfig(ctx context.Context, client *kubernetes.Clientset, namespace, secretName string) (kubeconfig []byte, err error) {
	secret, err := client.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("kubeconfig secret %s/%s not found", namespace, secretName)
		}
		return nil, fmt.Errorf("failed to get kubeconfig secret %s/%s: %w", namespace, secretName, err)
	}
	kubeconfig, ok := secret.Data["kubeconfig"]
	if !ok {
		return nil, fmt.Errorf("secret %s missing kubeconfig", secretName)
	}
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
	/*
		This line parses YAML file into *clientcmdapi.config which contains clusters
		authinfos, contexts, currentcontext.
	*/
	cfg, err := clientcmd.Load(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}
	/*
		A kubeconfig can have mutiple contexts, we are extracting a context needed
		for our controller. The name of the context will be same as registryClient.Cluster.Name
		If it does not exist we throw an error
	*/
	_, ok := cfg.Contexts[contextName]
	if !ok {
		return nil, fmt.Errorf("context %s not found in kubeconfig", contextName)
	}
	/*
		we have to override *clientcmdapi.config so that it uses the contextName = registryClient.Cluster.Labels.contexts

	*/
	overrides := &clientcmd.ConfigOverrides{CurrentContext: contextName}

	// We finally build a client config and return the client config's config as *rest.Config object

	clientCfg := clientcmd.NewDefaultClientConfig(*cfg, overrides)
	return clientCfg.ClientConfig()

}
