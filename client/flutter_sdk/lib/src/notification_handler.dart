import 'dart:io';

import 'package:flutter/foundation.dart';
import 'package:flutter_local_notifications/flutter_local_notifications.dart';

/// 本地通知处理器
class NotificationHandler {
  static NotificationHandler? _instance;
  final FlutterLocalNotificationsPlugin _plugin = FlutterLocalNotificationsPlugin();

  NotificationHandler._();

  static NotificationHandler get instance => _instance ??= NotificationHandler._();

  /// 初始化通知插件
  Future<void> init() async {
    if (!Platform.isAndroid && !Platform.isIOS) return;

    const androidSettings = AndroidInitializationSettings('@mipmap/ic_launcher');
    const iosSettings = DarwinInitializationSettings(
      requestAlertPermission: true,
      requestBadgePermission: true,
      requestSoundPermission: true,
    );

    const initSettings = InitializationSettings(
      android: androidSettings,
      iOS: iosSettings,
    );

    await _plugin.initialize(
      initSettings,
      onDidReceiveNotificationResponse: _onNotificationTapped,
    );
  }

  /// 检查通知权限是否已授权（Android 13+ / iOS）
  Future<bool> checkPermission() async {
    if (Platform.isAndroid) {
      final androidPlugin = _plugin.resolvePlatformSpecificImplementation<AndroidFlutterLocalNotificationsPlugin>();
      final enabled = await androidPlugin?.areNotificationsEnabled() ?? false;
      return enabled;
    }
    if (Platform.isIOS) {
      final iosPlugin = _plugin.resolvePlatformSpecificImplementation<IOSFlutterLocalNotificationsPlugin>();
      final settings = await iosPlugin?.requestPermissions(alert: true, badge: true, sound: true);
      return settings ?? false;
    }
    return false;
  }

  /// 请求通知权限（Android 13+ 需要显式请求）
  Future<bool> requestPermission() async {
    if (Platform.isAndroid) {
      final androidPlugin = _plugin.resolvePlatformSpecificImplementation<AndroidFlutterLocalNotificationsPlugin>();
      final granted = await androidPlugin?.requestNotificationsPermission() ?? false;
      debugPrint('[Notification] permission granted: $granted');
      return granted;
    }
    return true; // iOS 已在 init 时自动请求
  }

  /// 展示推送通知
  Future<void> showNotification({
    required String title,
    required String body,
    String? msgId,
  }) async {
    if (!Platform.isAndroid && !Platform.isIOS) return;

    const androidDetails = AndroidNotificationDetails(
      'push_platform_channel',
      '推送通知',
      channelDescription: '统一推送中台通知渠道',
      importance: Importance.high,
      priority: Priority.high,
    );

    const iosDetails = DarwinNotificationDetails();

    const details = NotificationDetails(
      android: androidDetails,
      iOS: iosDetails,
    );

    // msgId 的 hashCode 作为通知 ID，确保相同消息不重复
    final id = (msgId ?? title).hashCode;
    await _plugin.show(id.abs(), title, body, details);
  }

  void _onNotificationTapped(NotificationResponse response) {
    // 处理通知点击事件
  }
}
