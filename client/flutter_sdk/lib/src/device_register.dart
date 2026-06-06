import 'dart:convert';
import 'dart:io';

import 'package:device_info_plus/device_info_plus.dart';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';

/// 设备注册器
class DeviceRegister {
  final String serverUrl;
  final String appId;

  static const _tokenKey = 'push_platform_device_token';

  DeviceRegister({
    required this.serverUrl,
    required this.appId,
  });

  /// 获取或生成设备 token，并向服务器注册
  Future<String> register(String? userId) async {
    final prefs = await SharedPreferences.getInstance();
    String? token = prefs.getString(_tokenKey);

    token ??= await _generateToken();
    await prefs.setString(_tokenKey, token);

    final platform = _getPlatform();
    final response = await http.post(
      Uri.parse('$serverUrl/api/v1/device/register'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({
        'device_token': token,
        'user_id': userId,
        'platform': platform,
        'app_id': appId,
      }),
    );

    if (response.statusCode != 200) {
      throw Exception('Device registration failed: ${response.body}');
    }

    return token;
  }

  /// 获取已保存的设备 token
  Future<String?> getSavedToken() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(_tokenKey);
  }

  String _getPlatform() {
    if (Platform.isIOS) return 'ios';
    if (Platform.isAndroid) return 'android';
    // 鸿蒙暂归 android，后续适配
    return 'unknown';
  }

  Future<String> _generateToken() async {
    final deviceInfo = DeviceInfoPlugin();
    String identifier;

    if (Platform.isAndroid) {
      final info = await deviceInfo.androidInfo;
      identifier = '${info.brand}_${info.model}_${info.id}';
    } else if (Platform.isIOS) {
      final info = await deviceInfo.iosInfo;
      identifier = info.identifierForVendor ?? 'unknown';
    } else {
      identifier = DateTime.now().millisecondsSinceEpoch.toString();
    }

    return 'flutter_${appId}_${identifier}';
  }
}
