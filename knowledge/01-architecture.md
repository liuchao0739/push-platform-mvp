# 系统架构

## 整体架构图

```
业务方（后端服务）
    │
    ▼ POST /api/v1/push/send
┌─────────────┐
│  Push Server │ (Go + Gin)
│  :8080       │
└──────┬──────┘
       │
       ├──► 设备在线？──Yes──► WebSocket 直接投递 ──► Flutter App
       │                        └── ACK 回执
       │
       └──► 设备离线？──► 消息队列（Redis/Memory）
                              │
                              └──► 设备上线后补推
```

## 数据流（端到端）

1. Flutter App 启动 → `POST /api/v1/device/register` 注册设备 → 获取 WebSocket token
2. Flutter App → `ws://server/ws?token=xxx` 建立长连接 → 30s 心跳 ping
3. 业务方 → `POST /api/v1/push/send` 发送推送 → Server 查 Hub 中是否有该设备
4. 设备在线 → Hub.PushToDevice() 直接写入 WebSocket → App 收到 → 展示通知 + 回 ACK
5. 设备离线 → MQ.Publish() 写入消息队列 → 设备上线后消费补推

## 目录结构

```
server/
├── main.go          # 入口：初始化 DB/MQ/Hub、注册路由、启动服务
├── handlers/
│   └── handlers.go  # 3 个 API + 1 个统计接口
├── ws/
│   └── ws.go        # Hub 连接管理 + Client 读写泵
├── models/
│   └── models.go    # Device、PushMessage、请求/响应结构体
├── mq/
│   ├── memory.go    # 内存队列实现（本地开发）
│   └── redis.go     # Redis Streams 实现（生产环境）
└── sql/
    └── init.sql     # PostgreSQL 建表脚本

client/
├── flutter_sdk/     # Flutter SDK 库
│   └── lib/src/
│       ├── push_platform.dart     # SDK 单例入口
│       ├── ws_connection.dart     # WebSocket + 重连
│       ├── device_register.dart   # 设备注册
│       └── notification_handler.dart  # 本地通知
└── demo_app/        # Demo 应用
    └── lib/main.dart
```
