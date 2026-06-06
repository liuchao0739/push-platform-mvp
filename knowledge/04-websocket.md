# WebSocket 协议详情

## 连接建立

```
客户端：ws://host:8080/ws?token=flutter_demo_app_vivo_V2049A_xxx
服务端：HTTP 101 Switching Protocols → WebSocket 连接建立
```

## 消息格式

所有消息均为 JSON 文本帧。

### 客户端 → 服务端

#### 心跳 Ping
```json
{"type": "ping"}
```
- 间隔：30 秒
- 服务端收到后不做应用层响应（WebSocket 协议层自动处理 Pong）

#### 消息确认 ACK
```json
{"type": "ack", "msg_id": "msg_57d42b1a-ac8e-4c"}
```
- 客户端收到 push 消息后立即发送
- 服务端根据 msg_id 更新消息状态为 delivered

### 服务端 → 客户端

#### 推送消息
```json
{
  "type": "push",
  "msg_id": "msg_57d42b1a-ac8e-4c",
  "title": "测试推送",
  "body": "推送中台 MVP 验证成功",
  "app_id": "demo_app"
}
```

## 服务端实现细节

### Hub 模式
```
Hub { clients map[deviceToken]*Client }
  ├── register chan   ← serveWS() 创建 Client 后注册
  ├── unregister chan ← ReadPump 检测断线后注销
  └── broadcast chan  ← （预留）全量广播
```

### Client 结构
```go
type Client struct {
    Hub         *Hub
    Conn        *websocket.Conn
    Send        chan []byte      // 发送缓冲区（容量 256）
    DeviceToken string
}
```

### 读写泵（ReadPump / WritePump）
- **ReadPump**：goroutine 读取客户端消息（ping/ack），60s 无消息超时断开
- **WritePump**：goroutine 从 Send channel 读取并写入 WebSocket，54s 发送协议 Ping 帧
- 两个 goroutine 通过 Hub 的 register/unregister channel 协调生命周期

### PushToDevice 流程
```
handlers.SendPush()
  → json.Marshal(WSPushMessage)
  → hub.PushToDevice(deviceToken, data)
    → clients[deviceToken].Send <- data  // 非阻塞写入
    → WritePump 消费 Send channel
      → conn.WriteMessage(TextMessage, data)
```

## 客户端实现细节

### WSConnection 状态机
参见 `03-state-machine.md` 中「设备连接状态」章节。

### 重连触发条件
1. WebSocket 连接异常关闭（onError / onDone）
2. 心跳超时（30s 未收到服务端响应，通过 WebSocket 协议层 Ping/Pong 检测）
3. 应用从后台回到前台（预留，当前未实现）

### 重连时的设备注册
重连时不会重新调用 register API（token 已持久化），直接使用已保存的 token 建立 WebSocket 连接。
