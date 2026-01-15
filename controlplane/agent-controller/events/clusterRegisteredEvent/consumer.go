package clusterRegisteredEvent

import (
	"context"
	"encoding/json"
	"log"

	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/models/clusterRegistryModel"
	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/store/clusterRegistry"
	"github.com/redis/go-redis/v9"
)

type Consumer struct {
	rdb      *redis.Client
	stream   string
	group    string
	consumer string
	store    clusterRegistry.Store
}

func NewConsumer(rdb *redis.Client, stream, group, consumer string, store clusterRegistry.Store) *Consumer {
	return &Consumer{rdb: rdb, stream: stream, group: group, consumer: consumer, store: store}
}

func (c *Consumer) ensureGroup(ctx context.Context) {
	err := c.rdb.XGroupCreateMkStream(ctx, c.stream, c.group, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		panic("failed to create redis consumer group: " + err.Error())
	}
}

func (c *Consumer) Run(ctx context.Context) {
	// Ensure group function ensures that a consumer group is created (idempotent)
	c.ensureGroup(ctx)
	// If main context gets cancelled or parent context gets cancelled we return
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		// We are reading using XReadGroup with > streams and 10 count and 5 second block
		// The count and block are not based on SLO, but are such that read waits 5 seconds for the latest stream in each read
		// 10 count ensures, it recieves at max 10 messages, so CPU is not exhausted
		streams, err := c.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    c.group,
			Consumer: c.consumer,
			Streams:  []string{c.stream, ">"},
			Count:    10,
			Block:    5000,
		}).Result()
		if err != nil && err != redis.Nil {
			log.Printf("failed to read from redis stream: %v", err)
			continue
		}
		for _, stream := range streams {
			for _, message := range stream.Messages {
				raw, ok := message.Values["payload"]
				if !ok {
					_ = c.rdb.XAck(ctx, c.stream, c.group, message.ID)
					log.Printf("message missing payload: %v", message)
					continue
				}
				payloadStr, ok := raw.(string)
				if !ok {
					_ = c.rdb.XAck(ctx, c.stream, c.group, message.ID)
					log.Printf("message payload not string: %v", message)
					continue
				}
				var event ClusterRegisteredEvent
				err := json.Unmarshal([]byte(payloadStr), &event)
				if err != nil {
					_ = c.rdb.XAck(ctx, c.stream, c.group, message.ID)
					log.Printf("failed to unmarshal cluster registered event: %v", err)
					continue
				}
				if event.Type != "cluster.registered" {
					_ = c.rdb.XAck(ctx, c.stream, c.group, message.ID)
					continue
				}
				cluster := &clusterRegistryModel.ManagedCluster{
					ClusterName:   event.Cluster.Name,
					CredentialRef: event.Cluster.CredentialRef,
					Labels:        event.Cluster.Labels,
					RegisteredAt:  event.Timestamp,
					Source:        event.Source,
				}

				err = c.store.Insert(ctx, cluster)
				if err != nil {
					log.Printf("failed to insert/update managed cluster: %v", err)
					continue
				}

				// Acknowledge the message after successful processing
				_ = c.rdb.XAck(ctx, c.stream, c.group, message.ID)

			}
		}
	}
}
