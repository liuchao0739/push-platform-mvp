package ws

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client 单个 WebSocket 连接
type Client struct {
	Hub         *Hub
	Conn        *websocket.Conn
	Send        chan []byte
	DeviceToken string
	mu          sync.Mutex
}

// Hub 管理所有 WebSocket 连接
type Hub struct {
	clients    map[string]*Client // deviceToken -> Client
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			// 如果该设备已有连接，关闭旧连接
			if old, ok := h.clients[client.DeviceToken]; ok {
				close(old.Send)
				old.Conn.Close()
			}
			h.clients[client.DeviceToken] = client
			h.mu.Unlock()
			log.Printf("[Hub] device registered: %s, total: %d", client.DeviceToken, h.ClientCount())

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.DeviceToken]; ok {
				delete(h.clients, client.DeviceToken)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("[Hub] device unregistered: %s, total: %d", client.DeviceToken, h.ClientCount())

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.Send <- message:
				default:
					// 发送缓冲满，断开连接
					h.mu.RUnlock()
					h.mu.Lock()
					delete(h.clients, client.DeviceToken)
					close(client.Send)
					h.mu.Unlock()
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()

		case <-ticker.C:
			// 定时 ping 所有连接
			h.mu.RLock()
			for _, client := range h.clients {
				client.mu.Lock()
				client.Conn.WriteMessage(websocket.PingMessage, nil)
				client.mu.Unlock()
			}
			h.mu.RUnlock()
		}
	}
}

// Register 注册客户端连接
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister 注销客户端连接
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// PushToDevice 向指定设备推送消息
func (h *Hub) PushToDevice(deviceToken string, message []byte) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	client, ok := h.clients[deviceToken]
	if !ok {
		return false
	}
	select {
	case client.Send <- message:
		return true
	default:
		return false
	}
}

// ClientCount 当前在线设备数
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// ReadPump 读取客户端消息
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WS] read error: %v", err)
			}
			break
		}
		// 处理客户端 ack 消息等
		log.Printf("[WS] received from %s: %s", c.DeviceToken, string(message))
	}
}

// WritePump 向客户端发送消息
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			c.mu.Lock()
			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			c.mu.Unlock()
			if err != nil {
				log.Printf("[WS] write error: %v", err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			c.mu.Lock()
			err := c.Conn.WriteMessage(websocket.PingMessage, nil)
			c.mu.Unlock()
			if err != nil {
				return
			}
		}
	}
}
