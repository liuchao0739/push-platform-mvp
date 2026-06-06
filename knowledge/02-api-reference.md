# API 参考

## 1. 设备注册

```
POST /api/v1/device/register
Content-Type: application/json

请求体：
{
  "device_token": "flutter_demo_app_vivo_V2049A_xxx",  // 设备唯一标识，SDK 自动生成
  "user_id":      "user_001",                          // 用户 ID
  "platform":     "android",                           // ios / android / harmony / web
  "app_id":       "demo_app"                           // 应用标识
}

成功响应：
{
  "code": 0,
  "message": "ok",
  "data": {
    "device_token": "flutter_demo_app_vivo_V2049A_xxx",
    "platform": "android",
    "status": "active"
  }
}
```

### 字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| device_token | string | ✅ | 设备唯一标识，SDK 格式：`flutter_{app_id}_{brand}_{model}_{android_id}` |
| user_id | string | ❌ | 用户 ID，用于后续按用户推送 |
| platform | string | ✅ | 平台类型，枚举值：ios/android/harmony/web |
| app_id | string | ✅ | 应用标识，区分不同业务线 |

### 设计说明
- token 由 SDK 自动生成并持久化到 SharedPreferences，同一设备多次启动不变
- 重复注册（相同 device_token）会更新 user_id、platform、app_id，不会创建新记录

---

## 2. 发送推送

```
POST /api/v1/push/send
Content-Type: application/json

请求体：
{
  "app_id":             "demo_app",                          // 必填，应用标识
  "target_device_token": "flutter_demo_app_vivo_V2049A_xxx", // 可选，指定设备
  "target_user_id":      "user_001",                         // 可选，指定用户
  "title":              "测试推送",                           // 必填，推送标题
  "body":               "这是一条测试消息"                     // 必填，推送正文
}

成功响应（设备在线）：
{
  "code": 0,
  "message": "delivered via websocket",
  "data": {
    "msg_id": "msg_57d42b1a-ac8e-4c",
    "status": "delivered"
  }
}

成功响应（设备离线）：
{
  "code": 0,
  "message": "queued",
  "data": {
    "msg_id": "msg_57d42b1a-ac8e-4c",
    "status": "pending"
  }
}
```

### 投递逻辑
1. 优先尝试 WebSocket 直接投递（查 Hub 中是否有该 device_token）
2. 设备在线 → `status: delivered`，消息直接通过 WS 发给客户端
3. 设备离线 → `status: pending`，消息写入消息队列，等待设备上线补推
4. 同时写入数据库，记录全量消息历史

### target_device_token vs target_user_id
- `target_device_token`：精确投递到指定设备
- `target_user_id`：（预留）按用户推送，需要先查该用户的所有设备
- 当前 MVP 版本只实现了 device_token 级别的推送

---

## 3. 服务器状态

```
GET /api/v1/stats

响应：
{
  "code": 0,
  "data": {
    "online_devices": 1,   // 当前 WebSocket 在线设备数
    "total_devices": 5,    // 历史注册设备总数
    "total_messages": 23   // 历史消息总数
  }
}
```

---

## 4. WebSocket 连接

```
ws://host:8080/ws?token={device_token}

客户端 → 服务端消息：
{"type":"ping"}                        // 心跳（30s 间隔）
{"type":"ack","msg_id":"msg_xxx"}      // 消息确认

服务端 → 客户端消息：
{"type":"push","msg_id":"msg_xxx","title":"标题","body":"内容","app_id":"demo_app"}
```

### 连接参数
| 参数 | 必填 | 说明 |
|------|------|------|
| token | ✅ | 设备注册时获取的 device_token，用于关联设备身份 |

### 心跳机制
- 客户端每 30s 发送 `{"type":"ping"}`
- 服务端每 54s 发送 WebSocket 协议层的 Ping 帧
- 60s 无响应视为断线，服务端注销该设备
- 客户端断线后自动按指数退避重连
