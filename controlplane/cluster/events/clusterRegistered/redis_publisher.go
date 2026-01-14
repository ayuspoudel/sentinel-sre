package clusterRegistered

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ayuspoudel/sentinel-sre/controlplane/cluster/model"
	"github.com/redis/go-redis/v9"
)

type RedisPublisher struct {
	client     *redis.Client
	streamName string
}

/*
	Author: @ayuspoudel
	This publishes cluster intent events emitted by cluster registry.
	It is best-effort and non-blocking.
*/

func NewRedisPublisher(client *redis.Client, streamName string) *RedisPublisher {
	return &RedisPublisher{client: client, streamName: streamName}
}

func (p *RedisPublisher) PublishClusterRegistered(ctx context.Context, c *model.Cluster) error {
	if c == nil {
		return fmt.Errorf("nil cluster status")
	}
	clusterData := FromCluster(c)
	event := NewClusterDataEvent(clusterData)
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal cluster registered event: %w", err)
	}
	args := &redis.XAddArgs{Stream: p.streamName, ID: "*", Values: map[string]interface{}{
		"type":      event.Type,
		"source":    event.Source,
		"payload":   string(payload),
		"schema":    event.SchemaVersion,
		"timestamp": event.Timestamp.Format(time.RFC3339),
	},
	}
	err = p.client.XAdd(ctx, args).Err()
	if err != nil {
		return fmt.Errorf("failed to publish cluster registered event to redis: %w", err)
	}
	return nil

}
