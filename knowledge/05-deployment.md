# 部署与环境

## 本地开发（零依赖）

```bash
cd server
go run main.go
# 默认使用 SQLite + 内存队列，无需安装任何外部服务
# 启动后监听 :8080
```

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `DB_DRIVER` | `sqlite` | 数据库驱动，可选 `postgres` |
| `DB_PATH` | `push_platform.db` | SQLite 数据库文件路径 |
| `MQ_DRIVER` | `memory` | 消息队列驱动，可选 `redis` |
| `SERVER_PORT` | `:8080` | 服务监听端口 |

## Docker 部署

```bash
docker compose up -d
# 启动 PostgreSQL + Redis + Go Server
```

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `DB_DRIVER` | `postgres` | Docker 环境自动切 PostgreSQL |
| `DB_HOST` | `localhost` | PostgreSQL 地址 |
| `DB_PORT` | `5432` | PostgreSQL 端口 |
| `DB_USER` | `push_admin` | 数据库用户 |
| `DB_PASSWORD` | `push_pass_2026` | 数据库密码 |
| `DB_NAME` | `push_platform` | 数据库名 |
| `MQ_DRIVER` | `redis` | Docker 环境自动切 Redis |
| `REDIS_HOST` | `localhost` | Redis 地址 |
| `REDIS_PORT` | `6379` | Redis 端口 |

## 环境切换原理

`main.go` 中通过 `env()` 函数读取环境变量，fallback 到默认值：
- `DB_DRIVER=sqlite` → 使用 SQLite，数据库文件本地存储
- `DB_DRIVER=postgres` → 使用 PostgreSQL，通过 DSN 连接
- `MQ_DRIVER=memory` → 使用内存队列，进程内消费
- `MQ_DRIVER=redis` → 使用 Redis Streams，独立消费者 goroutine

## Flutter Demo App 运行

```bash
cd client/demo_app
flutter run
```

### Android 真机注意事项
1. **网络地址**：不能使用 `localhost`，需改为 Mac 局域网 IP
2. **通知权限**：Android 13+ 需在系统设置中开启通知权限，或 SDK 会自动请求
3. **desugaring**：`android/app/build.gradle.kts` 需启用 `isCoreLibraryDesugaringEnabled`

## 数据库表结构

### devices 表
| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER | 主键 |
| device_token | VARCHAR(255) | 设备唯一标识（唯一索引）|
| user_id | VARCHAR(128) | 用户 ID |
| platform | VARCHAR(32) | ios/android/harmony/web |
| app_id | VARCHAR(128) | 应用标识 |
| status | VARCHAR(16) | active/inactive |
| created_at | TIMESTAMP | 注册时间 |
| updated_at | TIMESTAMP | 更新时间 |

### push_messages 表
| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER | 主键 |
| msg_id | VARCHAR(64) | 消息唯一 ID（唯一索引）|
| app_id | VARCHAR(128) | 应用标识 |
| target_user_id | VARCHAR(128) | 目标用户 |
| target_device_token | VARCHAR(255) | 目标设备 |
| title | VARCHAR(256) | 推送标题 |
| body | TEXT | 推送正文 |
| status | VARCHAR(16) | pending/sent/delivered/failed |
| created_at | TIMESTAMP | 创建时间 |
| delivered_at | TIMESTAMP | 送达时间 |
