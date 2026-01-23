package kube

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeClient struct {
	client *kubernetes.Clientset
}

func NewKube() (*KubeClient, error) {
	client, err := newKubeClient()
	if err != nil {
		return nil, err
	}
	return &KubeClient{client: client}, nil
}

func newKubeClient() (*kubernetes.Clientset, error) {
	// Try to load in cluster config when running as a pod
	cfg, err := rest.InClusterConfig()
	if err == nil {
		return kubernetes.NewForConfig(cfg)
	}
	// If not running as a pod fetch kubeconfig from local (works on laptop, over docker while testing)
	// This reads from ~/.kube/config
	cfg, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(cfg)
}
