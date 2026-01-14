package clusterRuntimeStatusEvents

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/store/clusterRuntime"
	"github.com/redis/go-redis/v9"
)

type Consumer struct {
	rdb      *redis.Client
	stream   string
	group    string
	consumer string
	store    clusterRuntime.Store
}

func NewConsumer(rdb *redis.Client, stream string, group string, consumer string, store clusterRuntime.Store) *Consumer {
	return &Consumer{rdb: rdb, stream: stream, group: group, consumer: consumer, store: store}
}

func (c *Consumer) ensureGroup(ctx context.Context) {
	/*
		Author: @ayuspoudel
		We have created group with 0, that means from the beginning, because group creation can happen after events have already existed
		Even if consumer starts late it will be able to reconstruct the full state.
	*/
	err := c.rdb.XGroupCreateMkStream(ctx, c.stream, c.group, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		log.Fatalf("failed to create redis consumer group: %v", err)
	}
}

func (c *Consumer) Run(ctx context.Context) {
	// Ensure group function ensures that a consumer group is created (idempotent)
	c.ensureGroup(ctx)
	// If main context gets cancelled or parent context gets cancelled we return
	for {
		select {
		case <-ctx.Done():
			log.Println("shutting down cluster runtime events consumer")
			return
		default:
		}
		// We are reading using XReadGroup with > streams and 10 count and 5 second block
		// The count and block are not based on SLO, but are such that read waits 5 seconds for the latest stream in each read
		// 10 count ensures, it recieves at max 10 messages, so CPU is not exhausted
		streams, err := c.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    c.group,
			Consumer: c.consumer,
			/*
				Author: @ayuspoudel
				We have created group with 0, that means from the beginning, because group creation can happen after events are published.
				But messages should not be read from 0, it should always be read from when the group last acked at what index.
				This ensures even if consumer crashes and restarts, it will continue to read from where it left.
				This is also one of the reasons we want to use redis.
				c.stream is sentinel.events
			*/
			Streams: []string{c.stream, ">"}, // Meaning of > : "give me new messages that have never been delivered to any consumer in this group"
			// Since redis tracks which messages have already recieved ACK, it is helpful for our consumer to ignore already acked messages
			Count: 10,
			Block: 5 * time.Second,
		}).Result()
		if err != nil && err != redis.Nil {
			log.Printf("failed to read from redis stream: %v", err)
			continue
		}

		// Since read supports reading multiple streams although len(streans) = 1, for us; we still have to loop over it
		for _, stream := range streams {
			// Messages contain multiple messages so we have to loop over them
			// We are converting it into our event struct then using the event struct's fields to store into db, using store.UpsertClusterRuntime method.
			for _, message := range stream.Messages {
				payloadRaw, ok := message.Values["payload"]
				if !ok {
					_ = c.rdb.XAck(ctx, c.stream, c.group, message.ID)
					continue
				}
				payload, ok := payloadRaw.(string)
				if !ok {
					_ = c.rdb.XAck(ctx, c.stream, c.group, message.ID)
					continue
				}

				var event ClusterStatusEvent
				err = json.Unmarshal([]byte(payload), &event)
				if err != nil {
					log.Printf("invalid event payload: %v", err)
					continue
				}
				if event.Type != "agent.cluster_status.updated" {
					_ = c.rdb.XAck(ctx, c.stream, c.group, message.ID)
					continue

				}
				err = c.store.UpsertClusterRuntime(ctx, event.ClusterName, event.Status.Reachable, event.Status.AuthValid, event.Status.AgentInstalled, event.Status.AgentHealthy,
					event.Status.AgentVersion, event.Status.AgentNamespace)
				if err != nil {
					log.Printf("failed to upsert cluster runtime status: %v", err)
					continue
				}
				_ = c.rdb.XAck(ctx, c.stream, c.group, message.ID)
			}
		}
	}
}
