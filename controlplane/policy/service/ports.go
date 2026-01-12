package service

import "context"

/*
	Author: @ayuspoudel
	These are ports. These are interface definintions which we allow such that we
	can support this `implicity interface satisfaction`. What we do is this is defined
	in the service layer, and implemented in the adapter layer. At runtime we create
	the AdapterLayer interfaces by providing the inputs they need and finally pass it onto
	the service layer.
	Ex:
		in main.go
		pgDB := setupDB()
		clusterRegistryAdapter := adapters.NewClusterRegistryAdapter(pgDB)
		controllerReaderAdapter := adapters.NewControllerReaderAdapter(pgDB)
		promClientAdapter := adapters.NewPrometheusClientAdapter(promConfig)
			|
			|
		registryService := service.NewRegistryService(store, clusterRegistryAdapter, controllerReaderAdapter, promClientAdapter)
*/

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
