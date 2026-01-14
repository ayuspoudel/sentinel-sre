package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/logging"
	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/status"
	"github.com/redis/go-redis/v9"
)

/*
Author: @ayuspoudel
This publishes events (authoritative cluster status) snapshots emitted by agent controller. It is
non-blocking to our reconciliation loop, idempotent-safe and best-effort. It uses redis streams.
*/
type RedisPublisher struct {
	client     *redis.Client
	streamName string
}

func NewRedisPublisher(client *redis.Client, streamName string) *RedisPublisher {
	return &RedisPublisher{
		client:     client,
		streamName: streamName,
	}
}

func (p *RedisPublisher) PublishClusterStatus(ctx context.Context, st *status.ClusterStatus) error {
	log := logging.From(ctx)
	if st == nil {
		return fmt.Errorf("nil cluster status")
	}
	runtimeStatus := FromClusterStatus(st)
	event := NewClusterStatusEvent(st.ClusterName, runtimeStatus)
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal cluster status event: %w", err)
	}

	args := &redis.XAddArgs{Stream: p.streamName, ID: "*", Values: map[string]interface{}{
		"type":      event.Type,
		"cluster":   event.ClusterName,
		"source":    event.Source,
		"payload":   string(payload),
		"schema":    event.SchemaVersion,
		"timestamp": event.TimeStamp.Format(time.RFC3339),
	},
	}
	err = p.client.XAdd(ctx, args).Err()
	if err != nil {
		log.Error("failed to publish cluster status event to redis", "error", err, "cluster", st.ClusterName)
		return fmt.Errorf("failed to publish cluster status event to redis: %w", err)
	}

	log.Info("published cluster status event to redis", "cluster", st.ClusterName, "stream", p.streamName)
	return nil
}
