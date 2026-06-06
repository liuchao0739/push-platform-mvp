class PushMessage {
  final String type;
  final String msgId;
  final String title;
  final String body;
  final String appId;

  PushMessage({
    required this.type,
    required this.msgId,
    this.title = '',
    this.body = '',
    this.appId = '',
  });

  factory PushMessage.fromJson(Map<String, dynamic> json) {
    return PushMessage(
      type: json['type'] as String? ?? '',
      msgId: json['msg_id'] as String? ?? '',
      title: json['title'] as String? ?? '',
      body: json['body'] as String? ?? '',
      appId: json['app_id'] as String? ?? '',
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'type': type,
      'msg_id': msgId,
      'title': title,
      'body': body,
      'app_id': appId,
    };
  }

  @override
  String toString() => 'PushMessage(type: $type, msgId: $msgId, title: $title)';
}
