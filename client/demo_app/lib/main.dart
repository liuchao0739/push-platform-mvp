import 'package:flutter/material.dart';
import 'package:push_platform_sdk/push_platform_sdk.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Push Platform Demo',
      theme: ThemeData(
        colorSchemeSeed: Colors.blue,
        useMaterial3: true,
      ),
      home: const PushDemoPage(),
    );
  }
}

class PushDemoPage extends StatefulWidget {
  const PushDemoPage({super.key});

  @override
  State<PushDemoPage> createState() => _PushDemoPageState();
}

class _PushDemoPageState extends State<PushDemoPage> {
  final String _serverUrl = 'http://192.168.156.241:8080';
  final String _appId = 'demo_app';
  String _status = '未连接';
  String? _deviceToken;
  final List<PushMessage> _messages = [];

  @override
  void initState() {
    super.initState();
    _initSDK();
  }

  Future<void> _initSDK() async {
    try {
      final sdk = await PushPlatform.init(
        serverUrl: _serverUrl,
        appId: _appId,
        userId: 'demo_user_001',
      );

      sdk.onPushReceived = (msg) {
        setState(() {
          _messages.insert(0, msg);
        });
      };

      sdk.onConnectionChanged = (status) {
        setState(() {
          _status = status.name;
        });
      };

      await sdk.connect();

      setState(() {
        _deviceToken = sdk.deviceToken;
        _status = sdk.connectionStatus.name;
      });
    } catch (e) {
      setState(() {
        _status = '连接失败: $e';
      });
    }
  }

  @override
  void dispose() {
    PushPlatform.instance.disconnect();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('推送中台 Demo'),
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
      ),
      body: Column(
        children: [
          _buildStatusCard(),
          const Divider(),
          Expanded(child: _buildMessageList()),
        ],
      ),
    );
  }

  Widget _buildStatusCard() {
    final isConnected = _status == 'connected';
    return Card(
      margin: const EdgeInsets.all(16),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(
                  isConnected ? Icons.cloud_done : Icons.cloud_off,
                  color: isConnected ? Colors.green : Colors.grey,
                ),
                const SizedBox(width: 8),
                Text('状态: $_status',
                    style: Theme.of(context).textTheme.titleMedium),
              ],
            ),
            const SizedBox(height: 8),
            Text('Device Token: ${_deviceToken ?? "未注册"}',
                style: Theme.of(context).textTheme.bodySmall),
            Text('Server: $_serverUrl',
                style: Theme.of(context).textTheme.bodySmall),
            Text('App ID: $_appId',
                style: Theme.of(context).textTheme.bodySmall),
          ],
        ),
      ),
    );
  }

  Widget _buildMessageList() {
    if (_messages.isEmpty) {
      return const Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.notifications_none, size: 48, color: Colors.grey),
            SizedBox(height: 8),
            Text('暂无推送消息', style: TextStyle(color: Colors.grey)),
            SizedBox(height: 4),
            Text('通过 API 发送推送测试：\ncurl -X POST http://localhost:8080/api/v1/push/send',
                style: TextStyle(fontSize: 12, color: Colors.grey),
                textAlign: TextAlign.center),
          ],
        ),
      );
    }

    return ListView.builder(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      itemCount: _messages.length,
      itemBuilder: (context, index) {
        final msg = _messages[index];
        return Card(
          child: ListTile(
            leading: const Icon(Icons.notifications),
            title: Text(msg.title),
            subtitle: Text(msg.body),
            trailing: Text(
              msg.msgId.substring(0, msg.msgId.length > 12 ? 12 : msg.msgId.length),
              style: const TextStyle(fontSize: 10, color: Colors.grey),
            ),
          ),
        );
      },
    );
  }
}
