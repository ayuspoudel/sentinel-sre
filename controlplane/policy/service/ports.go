package service

import "context"

type ClusterRegistry interface {
	ClusterExists(ctx context.Context, name string) (bool, error)
}

type ControllerReader interface {
	ClusterReachable(ctx context.Context, cluster string) (bool, error)
	NamespaceExists(ctx context.Context, cluster, namespace string) (bool, error)
	AgentInstalled(ctx context.Context, cluster string) (bool, error)
	AgentHealthy(ctx context.Context, cluster string) (bool, error)
}

type PrometheusClient interface {
	Query(ctx context.Context, query string) error
}
