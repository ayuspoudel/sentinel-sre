package clusterRuntime

import "context"

/*
	Author: @ayuspoudel
	This storage is to consume events subscribed by policy registry.
	Events are published in the streamer which needs to be stored in
	db.
*/

type Store interface {
	ClusterReachable(ctx context.Context, cluster string) (bool, error)
	AgentInstalled(ctx context.Context, cluster string) (bool, error)
	AgentHealthy(ctx context.Context, cluster string) (bool, error)
	AgentNamespace(ctx context.Context, cluster string) (string, error)
	UpsertClusterRuntime(
		ctx context.Context,
		cluster string,
		reachable bool,
		authValid bool,
		agentInstalled bool,
		agentHealthy bool,
		agentVersion string,
		agentNamespace string,
	) error
}
