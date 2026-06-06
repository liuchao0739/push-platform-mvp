package mq

import (
	"encoding/json"
	"log"

	"push-platform/models"
)

// MessageQueue 消息队列接口
type MessageQueue interface {
	Publish(msg *models.WSPushMessage) error
}

// MemoryQueue 内存消息队列（本地开发用，零依赖）
type MemoryQueue struct {
	hub    interface{ PushToDevice(string, []byte) bool }
	queue  chan *models.WSPushMessage
}

// NewMemoryQueue 创建内存队列
func NewMemoryQueue(hub interface{ PushToDevice(string, []byte) bool }) *MemoryQueue {
	return &MemoryQueue{
		hub:   hub,
		queue: make(chan *models.WSPushMessage, 256),
	}
}

// Start 启动消费者
func (m *MemoryQueue) Start() {
	log.Println("[MQ] memory queue consumer started")
	for msg := range m.queue {
		pushData, err := json.Marshal(msg)
		if err != nil {
			log.Printf("[MQ] marshal error: %v", err)
			continue
		}

		// 广播给所有在线设备
		delivered := m.hub.PushToDevice(msg.AppID, pushData)
		status := "delivered"
		if !delivered {
			status = "pending"
		}
		log.Printf("[MQ] push result: msg_id=%s, status=%s", msg.MsgID, status)
	}
}

// Publish 发布消息到内存队列
func (m *MemoryQueue) Publish(msg *models.WSPushMessage) error {
	m.queue <- msg
	return nil
}
