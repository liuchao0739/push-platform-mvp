import 'dart:async';
import 'dart:convert';
import 'dart:math';

import 'package:flutter/foundation.dart';
import 'package:web_socket_channel/web_socket_channel.dart';

/// WebSocket 连接状态
enum WSStatus {
  disconnected,
  connecting,
  connected,
  reconnecting,
}

/// WebSocket 连接管理 + 心跳重连
class WSConnection {
  final String serverUrl;
  final String deviceToken;
  final void Function(Map<String, dynamic>) onMessage;

  WebSocketChannel? _channel;
  Timer? _heartbeatTimer;
  Timer? _reconnectTimer;
  StreamSubscription? _subscription;

  WSStatus _status = WSStatus.disconnected;
  int _reconnectAttempts = 0;
  static const int _maxReconnectDelay = 30; // 秒

  /// 连接状态变化回调
  void Function(WSStatus status)? onStatusChanged;

  WSConnection({
    required this.serverUrl,
    required this.deviceToken,
    required this.onMessage,
    this.onStatusChanged,
  });

  WSStatus get status => _status;

  /// 连接 WebSocket
  Future<void> connect() async {
    if (_status == WSStatus.connected || _status == WSStatus.connecting) {
      return;
    }

    _setStatus(WSStatus.connecting);

    try {
      final uri = Uri.parse('$serverUrl?token=$deviceToken');
      _channel = WebSocketChannel.connect(uri);

      _subscription = _channel!.stream.listen(
        _onData,
        onError: _onError,
        onDone: _onDone,
        cancelOnError: false,
      );

      _setStatus(WSStatus.connected);
      _reconnectAttempts = 0;
      _startHeartbeat();
    } catch (e) {
      _onError(e);
    }
  }

  /// 断开连接
  void disconnect() {
    _stopHeartbeat();
    _cancelReconnect();
    _subscription?.cancel();
    _channel?.sink.close();
    _channel = null;
    _subscription = null;
    _setStatus(WSStatus.disconnected);
  }

  /// 发送消息
  void send(Map<String, dynamic> data) {
    if (_status != WSStatus.connected || _channel == null) return;
    _channel!.sink.add(jsonEncode(data));
  }

  void _onData(dynamic message) {
    try {
      final data = jsonDecode(message as String) as Map<String, dynamic>;
      final raw = message.toString();
      debugPrint('[WS] received: ${raw.length > 200 ? raw.substring(0, 200) : raw}');
      onMessage(data);
    } catch (e) {
      debugPrint('[WS] parse error: $e, raw: $message');
    }
  }

  void _onError(dynamic error) {
    _scheduleReconnect();
  }

  void _onDone() {
    if (_status != WSStatus.disconnected) {
      _scheduleReconnect();
    }
  }

  /// 心跳：每 30 秒发送 ping
  void _startHeartbeat() {
    _stopHeartbeat();
    _heartbeatTimer = Timer.periodic(
      const Duration(seconds: 30),
      (_) => send({'type': 'ping'}),
    );
  }

  void _stopHeartbeat() {
    _heartbeatTimer?.cancel();
    _heartbeatTimer = null;
  }

  /// 指数退避重连：1s → 2s → 4s → 8s → 16s → 30s
  void _scheduleReconnect() {
    _stopHeartbeat();
    _setStatus(WSStatus.reconnecting);

    final delay = min(
      pow(2, _reconnectAttempts).toInt(),
      _maxReconnectDelay,
    );
    _reconnectAttempts++;

    _reconnectTimer = Timer(Duration(seconds: delay), () {
      connect();
    });
  }

  void _cancelReconnect() {
    _reconnectTimer?.cancel();
    _reconnectTimer = null;
    _reconnectAttempts = 0;
  }

  void _setStatus(WSStatus status) {
    _status = status;
    onStatusChanged?.call(status);
  }
}
