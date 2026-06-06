import 'package:flutter/foundation.dart';

import 'device_register.dart';
import 'models/push_message.dart';
import 'notification_handler.dart';
import 'ws_connection.dart';

/// 推送中台 Flutter SDK 入口
class PushPlatform {
  static PushPlatform? _instance;

  late final WSConnection _ws;
  late final DeviceRegister _register;
  late final NotificationHandler _notification;

  final String serverUrl;
  final String appId;
  String? _deviceToken;
  String? _userId;

  /// 收到推送消息回调
  void Function(PushMessage message)? onPushReceived;

  /// 连接状态变化回调
  void Function(WSStatus status)? onConnectionChanged;

  PushPlatform._({
    required this.serverUrl,
    required this.appId,
  });

  /// 获取 SDK 单例
  static PushPlatform get instance {
    if (_instance == null) {
      throw StateError('PushPlatform not initialized. Call PushPlatform.init() first.');
    }
    return _instance!;
  }

  /// 初始化 SDK
  static Future<PushPlatform> init({
    required String serverUrl,
    required String appId,
    String? userId,
  }) async {
    if (_instance != null) return _instance!;

    final sdk = PushPlatform._(serverUrl: serverUrl, appId: appId);
    sdk._userId = userId;

    // 初始化通知
    sdk._notification = NotificationHandler.instance;
    await sdk._notification.init();

    // 检查并请求通知权限（Android 13+ 需要显式请求）
    final hasPermission = await sdk._notification.checkPermission();
    if (!hasPermission) {
      await sdk._notification.requestPermission();
    }

    // 初始化设备注册
    sdk._register = DeviceRegister(
      serverUrl: serverUrl,
      appId: appId,
    );

    _instance = sdk;
    return sdk;
  }

  /// 注册设备并建立 WebSocket 连接
  Future<void> connect() async {
    // 1. 注册设备
    _deviceToken = await _register.register(_userId);

    // 2. 建立 WebSocket 连接
    final wsUrl = serverUrl.replaceFirst('http', 'ws');
    _ws = WSConnection(
      serverUrl: '$wsUrl/ws',
      deviceToken: _deviceToken!,
      onMessage: _handleMessage,
      onStatusChanged: (status) {
        onConnectionChanged?.call(status);
      },
    );

    await _ws.connect();
  }

  /// 断开连接
  void disconnect() {
    _ws.disconnect();
  }

  /// 发送 ACK 确认
  void ack(String msgId) {
    _ws.send({'type': 'ack', 'msg_id': msgId});
  }

  void _handleMessage(Map<String, dynamic> data) {
    debugPrint('[SDK] handleMessage: type=${data['type']}, msg_id=${data['msg_id']}');
    final msg = PushMessage.fromJson(data);

    switch (msg.type) {
      case 'push':
        debugPrint('[SDK] push received: title=${msg.title}, body=${msg.body}');
        onPushReceived?.call(msg);
        // 自动展示本地通知
        _notification.showNotification(
          title: msg.title,
          body: msg.body,
          msgId: msg.msgId,
        );
        // 自动发送 ACK
        ack(msg.msgId);
        break;
      case 'pong':
        // 心跳响应，忽略
        break;
      default:
        debugPrint('[SDK] Unknown message type: ${msg.type}');
    }
  }

  /// 获取当前连接状态
  WSStatus get connectionStatus => _ws.status;

  /// 获取设备 token
  String? get deviceToken => _deviceToken;
}
