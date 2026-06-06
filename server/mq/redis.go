package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"push-platform/models"
	"push-platform/ws"

	"github.com/go-redis/redis/v8"
)

const (
	streamName    = "push:messages"
	consumerGroup = "push-server"
)

// RedisQueue Redis Streams 消息队列（生产环境用）
type RedisQueue struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisQueue 创建 Redis 队列
func NewRedisQueue(host, port string) *RedisQueue {
	addr := fmt.Sprintf("%s:%s", host, port)
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   0,
	})
	ctx := context.Background()

	// 确保消费者组存在
	err := rdb.XGroupCreateMkStream(ctx, streamName, consumerGroup, "0").Err()
	if err != nil {
		log.Printf("[Redis] consumer group may already exist: %v", err)
	} else {
		log.Printf("[Redis] created consumer group %s on stream %s", consumerGroup, streamName)
	}

	return &RedisQueue{client: rdb, ctx: ctx}
}

// Publish 发布消息到 Redis Stream
func (r *RedisQueue) Publish(msg *models.WSPushMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	id, err := r.client.XAdd(r.ctx, &redis.XAddArgs{
		Stream: streamName,
		Values: map[string]interface{}{"body": string(data)},
	}).Result()
	if err != nil {
		return fmt.Errorf("xadd error: %w", err)
	}

	log.Printf("[Redis] message published, stream_id=%s, msg_id=%s", id, msg.MsgID)
	return nil
}

// GetClient 返回底层 redis.Client
func (r *RedisQueue) GetClient() *redis.Client {
	return r.client
}

// StartRedisConsumer 启动 Redis Stream 消费者
func StartRedisConsumer(rdb *RedisQueue, hub *ws.Hub) {
	consumerName := env("CONSUMER_NAME", "push-worker-1")
	log.Printf("[Redis] starting consumer %s...", consumerName)

	for {
		streams, err := rdb.client.XReadGroup(rdb.ctx, &redis.XReadGroupArgs{
			Group:    consumerGroup,
			Consumer: consumerName,
			Streams:  []string{streamName, ">"},
			Count:    10,
			Block:    5 * time.Second,
		}).Result()

		if err != nil {
			if err != redis.Nil {
				log.Printf("[Redis] xreadgroup error: %v", err)
			}
			continue
		}

		for _, stream := range streams {
			for _, message := range stream.Messages {
				body, ok := message.Values["body"].(string)
				if !ok {
					continue
				}

				var wsMsg models.WSPushMessage
				if err := json.Unmarshal([]byte(body), &wsMsg); err != nil {
					log.Printf("[Redis] unmarshal error: %v", err)
					continue
				}

				pushData, _ := json.Marshal(wsMsg)
				hub.PushToDevice(wsMsg.AppID, pushData)
				rdb.client.XAck(rdb.ctx, streamName, consumerGroup, message.ID)
			}
		}
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
