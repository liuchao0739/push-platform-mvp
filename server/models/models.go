package models

import "time"

// Device 设备注册信息
type Device struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	DeviceToken string    `gorm:"uniqueIndex;size:255;not null" json:"device_token"`
	UserID      string    `gorm:"size:128" json:"user_id"`
	Platform    string    `gorm:"size:32;not null;default:unknown" json:"platform"` // ios / android / harmony / web
	AppID       string    `gorm:"size:128;not null" json:"app_id"`
	Status      string    `gorm:"size:16;not null;default:active" json:"status"` // active / inactive
	CreatedAt   time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

// PushMessage 推送消息记录
type PushMessage struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	MsgID            string     `gorm:"uniqueIndex;size:64;not null" json:"msg_id"`
	AppID            string     `gorm:"size:128;not null" json:"app_id"`
	TargetUserID     string     `gorm:"size:128" json:"target_user_id"`
	TargetDeviceToken string    `gorm:"size:255" json:"target_device_token"`
	Title            string     `gorm:"size:256" json:"title"`
	Body             string     `gorm:"type:text" json:"body"`
	Status           string     `gorm:"size:16;not null;default:pending" json:"status"` // pending / sent / delivered / failed
	CreatedAt        time.Time  `gorm:"not null;autoCreateTime" json:"created_at"`
	DeliveredAt      *time.Time `json:"delivered_at"`
}

// RegisterRequest 设备注册请求
type RegisterRequest struct {
	DeviceToken string `json:"device_token" binding:"required"`
	UserID      string `json:"user_id"`
	Platform    string `json:"platform" binding:"required"` // ios / android / harmony / web
	AppID       string `json:"app_id" binding:"required"`
}

// SendPushRequest 推送发送请求
type SendPushRequest struct {
	AppID            string `json:"app_id" binding:"required"`
	TargetUserID     string `json:"target_user_id"`
	TargetDeviceToken string `json:"target_device_token"`
	Title            string `json:"title" binding:"required"`
	Body             string `json:"body" binding:"required"`
}

// WSPushMessage WebSocket 推送消息格式
type WSPushMessage struct {
	Type    string `json:"type"`    // push / ack / ping / pong
	MsgID   string `json:"msg_id"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	AppID   string `json:"app_id"`
}
