package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"push-platform/handlers"
	"push-platform/models"
	"push-platform/mq"
	"push-platform/ws"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
	// SQLite 本地模式（零依赖），Docker 部署时切 PostgreSQL
	dbDriver := env("DB_DRIVER", "sqlite")
	var db *gorm.DB
	var err error

	if dbDriver == "postgres" {
		dsn := buildDSN()
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	} else {
		dbPath := env("DB_PATH", "push_platform.db")
		db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	}

	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	db.AutoMigrate(&models.Device{}, &models.PushMessage{})

	hub := ws.NewHub()
	go hub.Run()

	// 消息队列：Redis（生产）或内存（本地开发）
	var messageQueue mq.MessageQueue
	mqDriver := env("MQ_DRIVER", "memory")
	if mqDriver == "redis" {
		messageQueue = mq.NewRedisQueue(
			env("REDIS_HOST", "localhost"),
			env("REDIS_PORT", "6379"),
		)
		go mq.StartRedisConsumer(messageQueue.(*mq.RedisQueue), hub)
	} else {
		messageQueue = mq.NewMemoryQueue(hub)
		go messageQueue.(*mq.MemoryQueue).Start()
	}

	r := gin.Default()
	r.Use(corsMiddleware())

	h := handlers.New(db, messageQueue, hub)

	api := r.Group("/api/v1")
	{
		api.POST("/device/register", h.RegisterDevice)
		api.POST("/push/send", h.SendPush)
		api.GET("/stats", h.Stats)
	}

	r.GET("/ws", func(c *gin.Context) {
		serveWS(hub, c.Writer, c.Request)
	})

	addr := env("SERVER_PORT", ":8080")
	log.Printf("Push Server starting on %s (db=%s, mq=%s)", addr, dbDriver, mqDriver)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func buildDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		env("DB_HOST", "localhost"),
		env("DB_PORT", "5432"),
		env("DB_USER", "push_admin"),
		env("DB_PASSWORD", "push_pass_2026"),
		env("DB_NAME", "push_platform"),
	)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func serveWS(hub *ws.Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	token := r.URL.Query().Get("token")
	if token == "" {
		token = "anonymous"
	}
	client := &ws.Client{
		Hub:         hub,
		Conn:        conn,
		Send:        make(chan []byte, 256),
		DeviceToken: token,
	}
	hub.Register(client)
	go client.WritePump()
	go client.ReadPump()
}
