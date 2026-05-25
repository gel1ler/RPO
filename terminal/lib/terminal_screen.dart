import 'dart:io' show Platform;

import 'package:flutter/material.dart';

import 'backend_client.dart';
import 'card_payload.dart';
import 'nfc_bridge.dart';

class TerminalScreen extends StatefulWidget {
  const TerminalScreen({super.key});

  @override
  State<TerminalScreen> createState() => _TerminalScreenState();
}

class _TerminalScreenState extends State<TerminalScreen> {
  final _baseCtrl = TextEditingController(text: 'https://127.0.0.1:8888/api/v1');
  final _serialCtrl = TextEditingController(text: 'TERM-MAC-001');
  final _amountCtrl = TextEditingController(text: '100');
  final _tripsCtrl = TextEditingController(text: '1');

  /// Номер карты по умолчанию — как в админке; для записи новой карты.
  final _cardNoCtrl = TextEditingController(text: '1DFC7D05');

  /// Key A hex с админки (12 символов).
  final _mfKeyHexCtrl = TextEditingController(text: 'FFFFFFFFFFFF');

  String _status = 'Готово.';
  bool _nfcBusy = false;
  CardPayload? _lastPayload;
  String _lastUid = '';
  Map<int, String> _keysById = {};
  List<Map<String, dynamic>> _keysList = [];

  BackendClient _api() =>
      BackendClient(baseUrl: _baseCtrl.text.trim(), allowBadCertificate: true);

  void _toast(String msg) {
    ScaffoldMessenger.maybeOf(context)?.showSnackBar(SnackBar(content: Text(msg)));
  }

  Future<void> _run(String label, Future<void> Function() action) async {
    setState(() {
      _nfcBusy = true;
      _status = label;
    });
    try {
      await action();
      if (!mounted) return;
      _toast('$label — ок');
    } catch (e, st) {
      debugPrint('$e\n$st');
      if (!mounted) return;
      _toast('$label: $e');
      setState(() => _status = e.toString());
    } finally {
      if (mounted) setState(() => _nfcBusy = false);
    }
  }

  Future<void> _loadKeys() async {
    final keys = await _api().loadKeys();
    final map = <int, String>{};
    final list = <Map<String, dynamic>>[];
    for (final raw in keys) {
      final id = (raw['id'] as num).toInt();
      final kv = raw['key_value'] as String? ?? '';
      map[id] = kv.trim();
      list.add(raw);
    }
    setState(() {
      _keysById = map;
      _keysList = list;
      _status = 'Загружено ключей: ${keys.length}';
    });
  }

  String _mifareKeyForPayload(CardPayload p) {
    final hex = _keysById[p.keyId] ?? _mfKeyHexCtrl.text.trim();
    return hex.trim();
  }

  Future<void> _readBalance() async {
    await _run('Считать карту', () async {
      final keyHex = _mfKeyHexCtrl.text.trim().replaceAll(RegExp(r'\s'), '');
      if (keyHex.length != 12) {
        throw StateError('MIFARE key: нужно ровно 12 hex-символов');
      }
      final r = NfcBridge.readCardPayload(keyHex.toUpperCase());
      final txt = r.json.trim().isEmpty ? '{}' : r.json;
      final p = CardPayload.parse(txt);
      setState(() {
        _lastUid = r.uid;
        _lastPayload = p;
        _status =
            'UID ${r.uid}: balance=${p.balance}, trips=${p.trips}, key_id=${p.keyId}';
      });
    });
  }

  Future<void> _initCardDefault() async {
    await _run('Инициализация карты', () async {
      final kid = _keysById.keys.isEmpty ? 1 : _keysById.keys.first;
      final keyHex = (_keysById[kid] ?? _mfKeyHexCtrl.text.trim())
          .trim()
          .replaceAll(RegExp(r'\s'), '')
          .toUpperCase();
      final p = CardPayload(
        v: 1,
        cardNumber: _cardNoCtrl.text.trim().toUpperCase(),
        balance: int.tryParse(_amountCtrl.text) ?? 1000,
        trips: int.tryParse(_tripsCtrl.text) ?? 0,
        keyId: kid,
      );
      NfcBridge.writeCardPayload(keyHex, p.toJsonString());
      final r = NfcBridge.readCardPayload(keyHex);
      final readBack =
          CardPayload.parse(r.json.trim().isEmpty ? '{}' : r.json);
      setState(() {
        _lastUid = r.uid;
        _lastPayload = readBack;
        _mfKeyHexCtrl.text = keyHex;
      });
    });
  }

  Future<void> _debitTerminal() async {
    final amt = int.tryParse(_amountCtrl.text) ?? 0;
    if (amt <= 0) {
      _toast('Сумма > 0');
      return;
    }
    await _run('Списание', () async {
      final keyHex = _mfKeyHexCtrl.text.trim().replaceAll(RegExp(r'\s'), '').toUpperCase();
      final snap = NfcBridge.readCardPayload(keyHex);
      final txt = snap.json.trim().isEmpty ? '{}' : snap.json;
      final payload = CardPayload.parse(txt);
      if (payload.cardNumber.trim().isEmpty) {
        throw StateError('Сначала выполните «Инициализировать карту»');
      }
      final resp = await _api().authorize(
        terminalSerial: _serialCtrl.text.trim(),
        cardNumber: payload.cardNumber,
        amount: amt,
      );
      final ok = resp['approved'] == true;
      final reason = resp['reason'] as String? ?? '';
      if (!ok) {
        throw StateError('Сервер отказал: $reason');
      }
      if (payload.balance < amt) {
        throw StateError(
            'На карте меньше, чем нужно списать (выставьте баланс карты как в SQLite).');
      }
      payload.balance -= amt;
      NfcBridge.writeCardPayload(
          _mifareKeyForPayload(payload).replaceAll(RegExp(r'\s'), '').toUpperCase(),
          payload.toJsonString());
      await _api().postTerminalEvent({
        'terminal_serial_number': _serialCtrl.text.trim(),
        'card_number': payload.cardNumber,
        'operation': 'debit_card',
        'amount': amt,
        'trips_delta': 0,
      });
      setState(() {
        _lastPayload = payload;
        _lastUid = snap.uid;
        _status = 'Списано $amt, balance на карте=${payload.balance}';
      });
    });
  }

  Future<void> _creditMoney() async {
    final amt = int.tryParse(_amountCtrl.text) ?? 0;
    if (amt <= 0) {
      _toast('Сумма > 0');
      return;
    }
    await _run('Пополнение счёта', () async {
      final keyHex = _mfKeyHexCtrl.text.trim().replaceAll(RegExp(r'\s'), '').toUpperCase();
      final snap = NfcBridge.readCardPayload(keyHex);
      final txt = snap.json.trim().isEmpty ? '{}' : snap.json;
      final payload = CardPayload.parse(txt);
      if (payload.cardNumber.trim().isEmpty) {
        throw StateError('Сначала инициализируйте карту');
      }
      payload.balance += amt;
      NfcBridge.writeCardPayload(_mifareKeyForPayload(payload).trim().replaceAll(RegExp(r'\s'), '').toUpperCase(),
          payload.toJsonString());
      await _api().postTerminalEvent({
        'terminal_serial_number': _serialCtrl.text.trim(),
        'card_number': payload.cardNumber,
        'operation': 'credit_balance',
        'amount': amt,
        'trips_delta': 0,
      });
      setState(() {
        _lastPayload = payload;
        _lastUid = snap.uid;
      });
    });
  }

  Future<void> _creditTrips() async {
    final d = int.tryParse(_tripsCtrl.text) ?? 0;
    if (d <= 0) {
      _toast('Поездок > 0');
      return;
    }
    await _run('Пополнение поездок', () async {
      final keyHex = _mfKeyHexCtrl.text.trim().replaceAll(RegExp(r'\s'), '').toUpperCase();
      final snap = NfcBridge.readCardPayload(keyHex);
      final txt = snap.json.trim().isEmpty ? '{}' : snap.json;
      final payload = CardPayload.parse(txt);
      if (payload.cardNumber.trim().isEmpty) {
        throw StateError('Сначала инициализируйте карту');
      }
      payload.trips += d;
      NfcBridge.writeCardPayload(_mifareKeyForPayload(payload).trim().replaceAll(RegExp(r'\s'), '').toUpperCase(),
          payload.toJsonString());
      await _api().postTerminalEvent({
        'terminal_serial_number': _serialCtrl.text.trim(),
        'card_number': payload.cardNumber,
        'operation': 'credit_trips',
        'amount': 0,
        'trips_delta': d,
      });
      setState(() {
        _lastPayload = payload;
        _lastUid = snap.uid;
      });
    });
  }

  @override
  void dispose() {
    _baseCtrl.dispose();
    _serialCtrl.dispose();
    _amountCtrl.dispose();
    _tripsCtrl.dispose();
    _cardNoCtrl.dispose();
    _mfKeyHexCtrl.dispose();
    super.dispose();
  }

  @override
  void initState() {
    super.initState();
    if (Platform.isMacOS) {
      final present = NfcBridge.readerPresent();
      _status =
          'PN532/USB: ${present ? "ридер доступен через libnfc" : "нет устройства в списке libnfc"}';
      setState(() {});
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('ЛР4 — NFC терминал')),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            TextField(
                controller: _baseCtrl,
                decoration: const InputDecoration(
                    labelText: 'HTTPS API база')),
            TextField(
                controller: _serialCtrl,
                decoration:
                    const InputDecoration(labelText: 'Серийный номер')),
            TextField(
                controller: _mfKeyHexCtrl,
                decoration:
                    const InputDecoration(labelText: 'Key A (12 hex)')),
            TextField(
                controller: _cardNoCtrl,
                decoration: const InputDecoration(
                    labelText: 'Номер карты (при записи новой карты)')),
            TextField(
                controller: _amountCtrl,
                keyboardType: TextInputType.number,
                decoration:
                    const InputDecoration(labelText: 'Сумма операции')),
            TextField(
                controller: _tripsCtrl,
                keyboardType: TextInputType.number,
                decoration: const InputDecoration(
                    labelText: 'Δ поездок (пополнение поездок)')),
            const Divider(height: 24),
            Text(_status, style: Theme.of(context).textTheme.bodyMedium),
            if (_lastUid.isNotEmpty) Text('UID: $_lastUid'),
            if (_lastPayload != null)
              Padding(
                padding: const EdgeInsets.only(top: 8),
                child: SelectableText(_lastPayload!.toJsonString()),
              ),
            const SizedBox(height: 16),
            Wrap(
              spacing: 12,
              runSpacing: 12,
              children: [
                FilledButton(
                  onPressed: _nfcBusy ? null : _loadKeys,
                  child: const Text('Ключи с сервера'),
                ),
                FilledButton.tonal(
                  onPressed: _nfcBusy ? null : _readBalance,
                  child: const Text('Баланс с карты'),
                ),
                FilledButton.tonal(
                  onPressed: _nfcBusy ? null : _initCardDefault,
                  child: const Text('Записать карту'),
                ),
                FilledButton(
                  onPressed: _nfcBusy ? null : _debitTerminal,
                  child: const Text('Списание'),
                ),
                FilledButton(
                  onPressed: _nfcBusy ? null : _creditMoney,
                  child: const Text('Пополнить деньги'),
                ),
                FilledButton(
                  onPressed: _nfcBusy ? null : _creditTrips,
                  child: const Text('Пополнить поездки'),
                ),
              ],
            ),
            if (_keysList.isNotEmpty) ...[
              const SizedBox(height: 24),
              Text(
                'Загруженные ключи (тап применить к полю)',
                style: Theme.of(context).textTheme.titleSmall,
              ),
              ..._keysList.map((k) {
                final id = k['id'];
                final lbl = k['label'] ?? '';
                final kv = (k['key_value'] ?? '').toString();
                final onlyHex = kv.replaceAll(RegExp(r'[^0-9A-Fa-f]'), '');
                return ListTile(
                  dense: true,
                  title: Text('$id — $lbl'),
                  subtitle: Text(kv),
                  onTap: () {
                    if (onlyHex.length >= 12) {
                      setState(() {
                        _mfKeyHexCtrl.text = onlyHex
                            .substring(onlyHex.length - 12)
                            .toUpperCase();
                      });
                    }
                  },
                );
              }),
            ],
          ],
        ),
      ),
    );
  }
}
