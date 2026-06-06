# Push Platform MVP

统一推送中台 MVP 原型 — AI（CodeBuddy）驱动的全流程开发验证

## 架构

```
┌─────────────────────────────────────┐
│          Go Push Server              │
│                                      │
│  POST /api/v1/push/send              │  ← 业务方调用
│  POST /api/v1/device/register        │  ← 客户端注册
│  GET  /ws?token=xxx                  │  ← WebSocket 长连接
│                                      │
│  ┌──────────┐  ┌──────────┐          │
│  │ Redis    │  │PostgreSQL│          │
│  │ Streams  │  │ 设备绑定  │          │
│  └──────────┘  └──────────┘          │
└──────────┬──────────────────────────┘
           │ WebSocket
    ┌──────▼──────┐
    │ Flutter SDK │
    │ • WS 连接   │
    │ • 心跳重连  │
    │ • 本地通知  │
    │ • 设备注册  │
    └─────────────┘
```

## 快速启动

```bash
# 启动基础设施
docker compose up -d postgres redis

# 启动 Go Server
cd server && go run main.go

# Flutter SDK 使用见 client/flutter_sdk/README.md
```

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端 | Go (gin + gorilla/websocket) |
| 消息队列 | Redis Streams |
| 数据库 | PostgreSQL |
| 客户端 | Flutter (Dart) |
| AI 工具 | CodeBuddy (glm-5.1) |

## 项目背景

公司「统一推送中台」项目方案设计阶段的 MVP 验证，目标是用 AI 将方案文档转化为可运行代码，验证技术选型可行性。
