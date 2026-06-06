# 排错指南

## 常见问题

### Q1：go run main.go 报 import 错误

**现象**：`could not import xxx (no required module provides package)`

**原因**：Go 依赖未下载

**解决**：
```bash
cd server
GOPROXY=https://goproxy.cn,direct go mod tidy
```

### Q2：消息发出去但 App 收不到（status: pending）

**原因**：设备不在线，或 device_token 不匹配

**排查步骤**：
1. `curl http://localhost:8080/api/v1/stats` 查看在线设备数
2. 确认 curl 发送推送时 `target_device_token` 与设备注册的 token 一致
3. 检查 App 界面显示的状态是否为 `connected`
4. 查看 Flutter 终端日志是否有 `[WS] received` / `[SDK] push received`

### Q3：消息发出去但 App 收不到（status: delivered）

**原因**：服务端已通过 WebSocket 发出，但客户端解析失败或通知权限不足

**排查步骤**：
1. 检查 Flutter 终端是否有 `[WS] parse error` 日志 → JSON 解析异常
2. 检查 Android 系统设置 → 通知 → push_platform_demo 是否已开启
3. 重新安装 App 后需手动开启通知权限（Android 13+）

### Q4：Android 编译报 desugaring 错误

**现象**：`Dependency ':flutter_local_notifications' requires core library desugaring`

**解决**：在 `android/app/build.gradle.kts` 中：
```kotlin
compileOptions {
    isCoreLibraryDesugaringEnabled = true  // 添加这一行
    sourceCompatibility = JavaVersion.VERSION_11
    targetCompatibility = JavaVersion.VERSION_11
}
defaultConfig {
    minSdk = 21  // 改为 21
}
dependencies {
    coreLibraryDesugaring("com.android.tools:desugar_jdk_libs:2.1.4")  // 添加依赖
}
```

### Q5：Android 真机连不上 localhost

**原因**：Android 真机上 `localhost` 指向手机自身，不是 Mac

**解决**：
1. Mac 终端执行 `ifconfig en0 | grep inet` 获取局域网 IP
2. 修改 `demo_app/lib/main.dart` 中的 `_serverUrl` 为 `http://{Mac IP}:8080`

### Q6：通知权限弹窗不出现

**原因**：首次安装时权限请求可能被系统忽略

**解决**：
- 手动：系统设置 → 应用 → push_platform_demo → 通知 → 开启
- 自动：SDK 在 `PushPlatform.init()` 时自动调用 `requestPermission()`

### Q7：WebSocket 频繁断线重连

**排查**：
1. 检查 Mac 和手机是否在同一 Wi-Fi 网络
2. 检查 Mac 防火墙是否阻止 8080 端口
3. 检查手机是否开启了省电模式（可能限制后台网络）
4. 观察服务端日志 `[WS] write error: i/o timeout` → 网络不稳定

### Q8：数据库文件过大

**原因**：消息记录无限增长

**当前状态**：MVP 未实现消息清理策略

**临时解决**：删除 `server/push_platform.db` 重新启动（数据全部丢失，仅开发环境使用）
