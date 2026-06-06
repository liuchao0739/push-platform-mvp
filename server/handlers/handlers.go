package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"push-platform/models"
	"push-platform/mq"
	"push-platform/ws"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Handler API 处理器
type Handler struct {
	db   *gorm.DB
	mq   mq.MessageQueue
	hub  *ws.Hub
}

// New 创建 Handler
func New(db *gorm.DB, messageQueue mq.MessageQueue, hub *ws.Hub) *Handler {
	return &Handler{db: db, mq: messageQueue, hub: hub}
}

// RegisterDevice 设备注册
// POST /api/v1/device/register
func (h *Handler) RegisterDevice(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	device := models.Device{
		DeviceToken: req.DeviceToken,
		UserID:      req.UserID,
		Platform:    req.Platform,
		AppID:       req.AppID,
		Status:      "active",
	}

	result := h.db.Where("device_token = ?", req.DeviceToken).
		Assign(map[string]interface{}{
			"user_id":    req.UserID,
			"platform":   req.Platform,
			"app_id":     req.AppID,
			"status":     "active",
			"updated_at": time.Now(),
		}).
		FirstOrCreate(&device)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	log.Printf("[API] device registered: token=%s, platform=%s, app_id=%s", req.DeviceToken, req.Platform, req.AppID)
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    device,
	})
}

// SendPush 发送推送
// POST /api/v1/push/send
func (h *Handler) SendPush(c *gin.Context) {
	var req models.SendPushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msgID := fmt.Sprintf("msg_%s", uuid.New().String()[:16])

	// 1. 写入数据库
	msg := models.PushMessage{
		MsgID:             msgID,
		AppID:             req.AppID,
		TargetUserID:      req.TargetUserID,
		TargetDeviceToken: req.TargetDeviceToken,
		Title:             req.Title,
		Body:              req.Body,
		Status:            "pending",
	}

	if err := h.db.Create(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 2. 构建 WebSocket 推送消息
	wsMsg := &models.WSPushMessage{
		Type:  "push",
		MsgID: msgID,
		Title: req.Title,
		Body:  req.Body,
		AppID: req.AppID,
	}

	// 3. 如果指定了 device token，直接尝试 WebSocket 推送
	if req.TargetDeviceToken != "" {
		pushData, _ := json.Marshal(wsMsg)
		if h.hub.PushToDevice(req.TargetDeviceToken, pushData) {
			now := time.Now()
			h.db.Model(&msg).Updates(map[string]interface{}{
				"status":       "delivered",
				"delivered_at": &now,
			})
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "delivered via websocket",
				"data": gin.H{
					"msg_id": msgID,
					"status": "delivered",
				},
			})
			return
		}
	}

	// 4. 设备不在线，写入消息队列等待后续投递
	if err := h.mq.Publish(wsMsg); err != nil {
		log.Printf("[API] mq publish error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "publish failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "queued",
		"data": gin.H{
			"msg_id": msgID,
			"status": "pending",
		},
	})
}

// Stats 服务器状态
// GET /api/v1/stats
func (h *Handler) Stats(c *gin.Context) {
	var deviceCount int64
	var messageCount int64
	h.db.Model(&models.Device{}).Count(&deviceCount)
	h.db.Model(&models.PushMessage{}).Count(&messageCount)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"data": gin.H{
			"online_devices": h.hub.ClientCount(),
			"total_devices":  deviceCount,
			"total_messages": messageCount,
		},
	})
}
