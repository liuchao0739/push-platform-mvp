# 统一推送中台 - 项目总览

## 项目定位

统一推送中台是公司的消息推送基础设施，目标是为各业务线（App、小程序、Web）提供统一的推送通道，解决当前各业务线各自对接厂商 SDK 导致的维护成本高、送达率不透明的问题。

## 核心模块

| 模块 | 职责 | 关键技术 |
|------|------|----------|
| Push Server | 接收推送请求、投递消息 | Go + Gin + WebSocket |
| WebSocket 网关 | 维持设备长连接、消息下发 | gorilla/websocket + Hub 模式 |
| 消息队列 | 削峰填谷、离线消息暂存 | 内存队列（开发）/ Redis Streams（生产）|
| Flutter SDK | 客户端接入层 | WebSocket 连接 + 心跳重连 + 本地通知 |

## 技术栈

- **后端**：Go 1.22 + Gin + gorilla/websocket + GORM
- **数据库**：SQLite（本地开发）/ PostgreSQL（生产环境）
- **消息队列**：内存队列（本地）/ Redis Streams（生产）
- **客户端 SDK**：Flutter/Dart + web_socket_channel
- **部署**：Docker Compose（PostgreSQL + Redis + Go Server）

## 核心设计决策

1. **MVP 不含厂商通道**：华为/小米/OPPO/vivo 厂商推送暂未接入，仅验证 WebSocket 自建通道
2. **双模式切换**：通过环境变量 `DB_DRIVER`、`MQ_DRIVER` 在本地零依赖和生产部署之间切换
3. **ACK 确认机制**：客户端收到消息后回执 `{"type":"ack","msg_id":"xxx"}`，服务端据此更新投递状态
4. **指数退避重连**：Flutter SDK 断线后按 1s→2s→4s→8s→16s→30s 递增重试间隔
