import 'dart:convert';
import 'dart:io';

typedef JsonMap = Map<String, dynamic>;

class BackendClient {
  BackendClient({
    required this.baseUrl,
    this.allowBadCertificate = true,
  }) {
    if (!baseUrl.endsWith('/')) {
      baseUrl = '$baseUrl/';
    }
  }

  /// Например `https://127.0.0.1:8888/api/v1`
  String baseUrl;
  bool allowBadCertificate;
  String? _token;

  Uri _uri(String path) {
    final tail = path.startsWith('/') ? path.substring(1) : path;
    return Uri.parse('$baseUrl$tail');
  }

  HttpClient _client() {
    final c = HttpClient();
    c.badCertificateCallback = (cert, host, port) => allowBadCertificate;
    return c;
  }

  Map<String, String> _jsonHeaders(int contentLength) => {
        'Content-Type': 'application/json',
        'Content-Length': contentLength.toString(),
        if (_token != null && _token!.isNotEmpty) 'Authorization': 'Bearer $_token',
      };

  /// Получить JWT терминала (POST /terminal/auth/login).
  Future<void> login({
    required String serialNumber,
    required String apiSecret,
  }) async {
    final body = utf8.encode(jsonEncode({
      'serial_number': serialNumber,
      'api_secret': apiSecret,
    }));

    final c = _client();
    try {
      final req = await c.postUrl(_uri('/terminal/auth/login'));
      req.headers.contentType = ContentType.json;
      req.headers.contentLength = body.length;
      req.add(body);
      final resp = await req.close();
      final text = await resp.transform(utf8.decoder).join();
      if (resp.statusCode != HttpStatus.ok) {
        throw HttpException(text, uri: req.uri);
      }
      final data = jsonDecode(text) as JsonMap;
      final token = data['token'] as String?;
      if (token == null || token.isEmpty) {
        throw const HttpException('login response missing token');
      }
      _token = token;
    } finally {
      c.close(force: true);
    }
  }

  Future<JsonMap> authorize({
    required String terminalSerial,
    required String cardNumber,
    required int amount,
  }) async {
    final body = utf8.encode(jsonEncode({
      'terminal_serial_number': terminalSerial,
      'card_number': cardNumber,
      'amount': amount,
    }));

    final c = _client();
    try {
      final req = await c.postUrl(_uri('/terminal/authorize'));
      for (final e in _jsonHeaders(body.length).entries) {
        req.headers.set(e.key, e.value);
      }
      req.add(body);
      final resp = await req.close();
      final text = await resp.transform(utf8.decoder).join();
      if (resp.statusCode != HttpStatus.ok) {
        throw HttpException(text, uri: req.uri);
      }
      return jsonDecode(text) as JsonMap;
    } finally {
      c.close(force: true);
    }
  }

  /// Регистрация карты в SQLite (создание или обновление баланса/key_id).
  Future<JsonMap> registerCard({
    required String terminalSerial,
    required String cardNumber,
    required int balance,
    required int keyId,
  }) async {
    final body = utf8.encode(jsonEncode({
      'terminal_serial_number': terminalSerial,
      'card_number': cardNumber,
      'balance': balance,
      'key_id': keyId,
    }));

    final c = _client();
    try {
      final req = await c.postUrl(_uri('/terminal/register-card'));
      for (final e in _jsonHeaders(body.length).entries) {
        req.headers.set(e.key, e.value);
      }
      req.add(body);
      final resp = await req.close();
      final text = await resp.transform(utf8.decoder).join();
      if (resp.statusCode != HttpStatus.ok) {
        throw HttpException(text, uri: req.uri);
      }
      return jsonDecode(text) as JsonMap;
    } finally {
      c.close(force: true);
    }
  }

  /// Синхронизация текущего состояния карты с БД после записи на чип.
  Future<JsonMap> syncCardState({
    required String terminalSerial,
    required String cardNumber,
    required int balance,
    required int keyId,
  }) {
    return registerCard(
      terminalSerial: terminalSerial,
      cardNumber: cardNumber,
      balance: balance,
      keyId: keyId,
    );
  }

  Future<List<JsonMap>> loadKeys() async {
    final c = _client();
    try {
      final req = await c.getUrl(_uri('/terminal/keys'));
      if (_token != null && _token!.isNotEmpty) {
        req.headers.set('Authorization', 'Bearer $_token');
      }
      final resp = await req.close();
      final text = await resp.transform(utf8.decoder).join();
      if (resp.statusCode != HttpStatus.ok) {
        throw HttpException(text, uri: req.uri);
      }
      return (jsonDecode(text) as List<dynamic>).cast<JsonMap>();
    } finally {
      c.close(force: true);
    }
  }

  /// Опубликовать операцию после записи на карту (не для authorize — он уже в БД).
  Future<JsonMap> postTerminalEvent(JsonMap payload) async {
    final body = utf8.encode(jsonEncode(payload));
    final c = _client();
    try {
      final req = await c.postUrl(_uri('/terminal/event'));
      for (final e in _jsonHeaders(body.length).entries) {
        req.headers.set(e.key, e.value);
      }
      req.add(body);
      final resp = await req.close();
      final text = await resp.transform(utf8.decoder).join();
      if (resp.statusCode != HttpStatus.created) {
        throw HttpException('${resp.statusCode} $text', uri: req.uri);
      }
      return jsonDecode(text) as JsonMap;
    } finally {
      c.close(force: true);
    }
  }
}
