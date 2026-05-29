/// Параметры терминала (не настраиваются из UI).
abstract final class TerminalConfig {
  static const apiBaseUrl = 'https://127.0.0.1:8888/api/v1';
  static const serialNumber = 'TERM-MAC-001';
  /// Секрет устройства (совпадает с APP_TERMINAL_API_SECRET в docker-compose).
  static const apiSecret = 'terminal-dev-secret';
  static const defaultMifareKeyHex = 'FFFFFFFFFFFF';
}
