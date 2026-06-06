# 合并审查 Agent 演示

## 场景：两个同事同时改了用户模块

### 模拟冲突

**同事 A（feature/sms-login）**：加了短信登录
```dart
// lib/api/user_api.dart
class UserApi {
  Future<String> loginByPassword(String phone, String password) async { ... }
  Future<String> loginBySms(String phone, String code) async { ... }  // A 新增
}
```

**同事 B（feature/password-policy）**：改了密码策略
```dart
// lib/api/user_api.dart
class UserApi {
  Future<String> loginByPassword(String phone, String password) async {
    if (password.length < 8) throw Exception('密码至少8位');  // B 修改了业务逻辑
    ...
  }
}
```

合并时 Git 标记冲突：
```
<<<<<<< feature/sms-login
  Future<String> loginBySms(String phone, String code) async { ... }
=======
  Future<String> loginByPassword(...) async {
    if (password.length < 8) throw Exception('密码至少8位');
    ...
  }
>>>>>>> feature/password-policy
```

---

## 直接问 AI（一次性）

> «这个冲突怎么解决？»

AI 可能回答：「保留 B 的密码校验，同时加上 A 的短信登录方法。」
→ **看起来没问题**，但你不知道：B 的密码校验改了业务规则，上线前需要测试确认。

---

## 合并审查 Agent（工作流）

### Agent 自动做的事

1. `git blame lib/api/user_api.dart` → 发现 A 和 B 都是最近 2 天改的
2. `git log feature/password-policy --oneline` → 看到 B 的 commit message: «重构密码校验，对齐安全规范 v2»
3. 读取项目规则 → 发现 `lib/api/` 文件属于「建议解决」级别
4. 读取项目规则 → B 的改动涉及**核心逻辑变更**，应该升为「人工介入」

### Agent 输出

```
## 合并审查报告

### 🟡 lib/api/user_api.dart
冲突类型: 同一文件新增 vs 修改

建议: 保留双方改动，但需注意：
  ⚠️ B 修改了密码校验逻辑（对齐安全规范 v2），建议：
     1. 确认测试用例已覆盖新密码规则
     2. 确认旧密码（<8位）的兼容策略

### 合并后校验
✅ 无残留冲突标记
✅ import 完整
⚠️ B 新增了 Exception 类型，需确认 lib/utils/exceptions.dart 已同步
```

---

## 核心差异

| | 直接问 AI | Agent/工作流 |
|---|---|---|
| 触发 | 手动贴代码 | 告诉你"合并 feature/sms-login → develop" |
| 上下文 | 只看冲突片段 | 读 git blame + commit message + 项目规则 |
| 一致性 | 每次回答可能不同 | 同一套规则，每次输出一致 |
| 红线 | 无 | pubspec.yaml 被覆盖→直接拦截 |
| 记录 | 无 | 每份报告可留存，追溯谁做了什么决策 |
