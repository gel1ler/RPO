import 'dart:io' show Platform;

import 'package:flutter/material.dart';

import 'backend_client.dart';
import 'card_payload.dart';
import 'card_payload_view.dart';
import 'nfc_bridge.dart';
import 'terminal_config.dart';

class TerminalScreen extends StatefulWidget {
  const TerminalScreen({super.key});

  @override
  State<TerminalScreen> createState() => _TerminalScreenState();
}

class _TerminalScreenState extends State<TerminalScreen>
    with TickerProviderStateMixin {
  static const _cardWaitDuration = Duration(seconds: 5);
  static const _kTabRegister = 0;
  static const _kTabInfo = 1;
  static const _kTabDebit = 2;
  static const _kTabCredit = 3;

  final _initBalanceCtrl = TextEditingController(text: '1000');
  final _debitAmountCtrl = TextEditingController(text: '100');
  final _creditAmountCtrl = TextEditingController(text: '100');

  final _api = BackendClient(
    baseUrl: TerminalConfig.apiBaseUrl,
    allowBadCertificate: true,
  );

  late final TabController _tabController;
  late final AnimationController _cardWaitController;

  String _status = 'Готово.';
  bool _cardWaitVisible = false;
  String? _busyOperation;
  bool _readerPresent = false;
  CardPayload? _displayedPayload;
  int? _cardShownOnTab;
  Map<int, String> _keysById = {};
  String? _keysLoadError;

  @override
  void dispose() {
    _tabController.removeListener(_onTabChanged);
    _tabController.dispose();
    _cardWaitController.dispose();
    _initBalanceCtrl.dispose();
    _debitAmountCtrl.dispose();
    _creditAmountCtrl.dispose();
    super.dispose();
  }

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 4, vsync: this);
    _tabController.addListener(_onTabChanged);
    _cardWaitController = AnimationController(
      vsync: this,
      duration: _cardWaitDuration,
    )..addListener(() {
        if (_cardWaitVisible && mounted) setState(() {});
      });
    _refreshReaderStatus();
    _loadKeysQuietly();
  }

  void _onTabChanged() {
    if (_tabController.indexIsChanging) return;
    _hideCardInfo();
  }

  void _hideCardInfo() {
    if (_displayedPayload == null && _cardShownOnTab == null) return;
    setState(() {
      _displayedPayload = null;
      _cardShownOnTab = null;
    });
  }

  void _revealCardInfo(CardPayload payload) {
    setState(() {
      _displayedPayload = payload;
      _cardShownOnTab = _tabController.index;
    });
  }

  void _refreshReaderStatus() {
    if (!Platform.isMacOS) return;
    setState(() => _readerPresent = NfcBridge.readerPresent());
  }

  Future<void> _loadKeysQuietly() async {
    try {
      final keys = await _api.loadKeys();
      final map = <int, String>{};
      for (final raw in keys) {
        final id = (raw['id'] as num).toInt();
        map[id] = (raw['key_value'] as String? ?? '').trim();
      }
      if (!mounted) return;
      setState(() {
        _keysById = map;
        _keysLoadError = null;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() => _keysLoadError = e.toString());
    }
  }

  String _extractMifareHex(String keyValue) {
    final onlyHex = keyValue.replaceAll(RegExp(r'[^0-9A-Fa-f]'), '');
    if (onlyHex.length >= 12) {
      return onlyHex.substring(onlyHex.length - 12).toUpperCase();
    }
    return TerminalConfig.defaultMifareKeyHex;
  }

  String get _defaultMifareKeyHex {
    if (_keysById.isEmpty) return TerminalConfig.defaultMifareKeyHex;
    return _extractMifareHex(_keysById[_keysById.keys.first]!);
  }

  int get _defaultKeyId =>
      _keysById.keys.isEmpty ? 1 : _keysById.keys.first;

  String _mifareKeyForPayload(CardPayload p) {
    final raw = _keysById[p.keyId];
    if (raw != null && raw.isNotEmpty) {
      return _extractMifareHex(raw);
    }
    return _defaultMifareKeyHex;
  }

  String _requireMifareKeyHex() {
    final hex = _defaultMifareKeyHex.replaceAll(RegExp(r'\s'), '');
    if (hex.length != 12) {
      throw StateError(
          'MIFARE key: нужно 12 hex-символов (проверьте ключи на сервере)');
    }
    return hex;
  }

  void _toast(String msg) {
    ScaffoldMessenger.maybeOf(context)?.showSnackBar(SnackBar(content: Text(msg)));
  }

  /// Ожидание карты (до 5 с в C) + анимированная шкала на UI.
  Future<T> _duringCardWait<T>(Future<T> Function() nfc) async {
    setState(() => _cardWaitVisible = true);
    _cardWaitController.reset();
    final animation = _cardWaitController.forward();
    try {
      return await nfc();
    } finally {
      if (mounted) {
        setState(() => _cardWaitVisible = false);
        _cardWaitController.stop();
        _cardWaitController.reset();
      }
      await animation.catchError((_) {});
    }
  }

  Future<void> _run(String label, Future<void> Function() action) async {
    setState(() {
      _busyOperation = label;
      _status = '$label — приложите карту к ридеру';
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
      if (mounted) {
        setState(() => _busyOperation = null);
        _refreshReaderStatus();
      }
    }
  }

  Future<void> _readBalance() async {
    await _run('Считать карту', () async {
      final keyHex = _requireMifareKeyHex();
      final r = await _duringCardWait(
        () => NfcBridge.readCardPayloadAsync(keyHex),
      );
      final txt = r.json.trim().isEmpty ? '{}' : r.json;
      final p = CardPayload.parse(txt);
      setState(() => _status =
          'Карта ${p.cardNumber.isNotEmpty ? p.cardNumber : r.uid}: balance=${p.balance}');
      _revealCardInfo(p);
    });
  }

  Future<void> _initCardDefault() async {
    await _run('Регистрация карты', () async {
      final keyHex = _requireMifareKeyHex();
      final snap = await _duringCardWait(
        () => NfcBridge.readCardPayloadAsync(keyHex),
      );
      if (snap.uid.isEmpty) {
        throw StateError('Не удалось прочитать UID — приложите карту');
      }
      final kid = _defaultKeyId;
      final p = CardPayload(
        v: 1,
        cardNumber: snap.uid,
        balance: int.tryParse(_initBalanceCtrl.text) ?? 1000,
        keyId: kid,
      );
      await _duringCardWait(
        () => NfcBridge.writeCardPayloadAsync(keyHex, p.toJsonString()),
      );
      final r = await _duringCardWait(
        () => NfcBridge.readCardPayloadAsync(keyHex),
      );
      final readBack =
          CardPayload.parse(r.json.trim().isEmpty ? '{}' : r.json);

      final backend = await _api.registerCard(
        terminalSerial: TerminalConfig.serialNumber,
        cardNumber: readBack.cardNumber,
        balance: readBack.balance,
        keyId: readBack.keyId,
      );
      final created = backend['created'] == true;
      final card = backend['card'] as Map<String, dynamic>?;
      final dbId = card?['id'];

      setState(() => _status = created
          ? 'Карта записана на чип и зарегистрирована в БД (id=$dbId)'
          : 'Карта на чипе обновлена, запись в БД синхронизирована (id=$dbId)');
      _revealCardInfo(readBack);
    });
  }

  Future<void> _debitTerminal() async {
    final amt = int.tryParse(_debitAmountCtrl.text) ?? 0;
    if (amt <= 0) {
      _toast('Сумма списания > 0');
      return;
    }
    await _run('Списание', () async {
      final keyHex = _requireMifareKeyHex();
      final snap = await _duringCardWait(
        () => NfcBridge.readCardPayloadAsync(keyHex),
      );
      final txt = snap.json.trim().isEmpty ? '{}' : snap.json;
      final payload = CardPayload.parse(txt);
      if (payload.cardNumber.trim().isEmpty) {
        throw StateError('Сначала зарегистрируйте карту');
      }
      final resp = await _api.authorize(
        terminalSerial: TerminalConfig.serialNumber,
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
            'На карте меньше, чем нужно списать (согласуйте баланс с записью в админке).');
      }
      payload.balance -= amt;
      await _duringCardWait(
        () => NfcBridge.writeCardPayloadAsync(
          _mifareKeyForPayload(payload),
          payload.toJsonString(),
        ),
      );
      await _api.postTerminalEvent({
        'terminal_serial_number': TerminalConfig.serialNumber,
        'card_number': payload.cardNumber,
        'operation': 'debit_card',
        'amount': amt,
      });
      setState(() => _status = 'Списано $amt, balance на карте=${payload.balance}');
      _revealCardInfo(payload);
    });
  }

  Future<void> _creditMoney() async {
    final amt = int.tryParse(_creditAmountCtrl.text) ?? 0;
    if (amt <= 0) {
      _toast('Сумма пополнения > 0');
      return;
    }
    await _run('Пополнение', () async {
      final keyHex = _requireMifareKeyHex();
      final snap = await _duringCardWait(
        () => NfcBridge.readCardPayloadAsync(keyHex),
      );
      final txt = snap.json.trim().isEmpty ? '{}' : snap.json;
      final payload = CardPayload.parse(txt);
      if (payload.cardNumber.trim().isEmpty) {
        throw StateError('Сначала зарегистрируйте карту');
      }
      payload.balance += amt;
      await _duringCardWait(
        () => NfcBridge.writeCardPayloadAsync(
          _mifareKeyForPayload(payload),
          payload.toJsonString(),
        ),
      );
      await _api.postTerminalEvent({
        'terminal_serial_number': TerminalConfig.serialNumber,
        'card_number': payload.cardNumber,
        'operation': 'credit_balance',
        'amount': amt,
      });
      setState(() => _status = 'Пополнено $amt, balance=${payload.balance}');
      _revealCardInfo(payload);
    });
  }

  Widget _statusDot(Color color) {
    return Container(
      width: 10,
      height: 10,
      decoration: BoxDecoration(color: color, shape: BoxShape.circle),
    );
  }

  Widget _terminalHeader(BuildContext context) {
    final theme = Theme.of(context);
    final readerOk = _readerPresent;
    final readerColor = readerOk ? Colors.green : Colors.red;
    final readerLabel = readerOk ? 'Ридер доступен' : 'Ридер недоступен';

    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Expanded(
          child: Text('NFC терминал', style: theme.textTheme.titleLarge),
        ),
        Column(
          crossAxisAlignment: CrossAxisAlignment.end,
          children: [
            Text('SN ${TerminalConfig.serialNumber}',
                style: theme.textTheme.bodySmall),
            const SizedBox(height: 2),
            Text('Key $_defaultMifareKeyHex',
                style: theme.textTheme.bodySmall
                    ?.copyWith(fontFamily: 'monospace')),
            if (_keysLoadError != null) ...[
              const SizedBox(height: 2),
              Text('Ключи: ошибка загрузки',
                  style: theme.textTheme.bodySmall
                      ?.copyWith(color: theme.colorScheme.error)),
            ],
            const SizedBox(height: 8),
            Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                _statusDot(readerColor),
                const SizedBox(width: 6),
                Text(readerLabel, style: theme.textTheme.bodySmall),
              ],
            ),
          ],
        ),
      ],
    );
  }

  Widget _statusBlock(BuildContext context) {
    final theme = Theme.of(context);
    final elapsed =
        (_cardWaitController.value * _cardWaitDuration.inMilliseconds / 1000)
            .clamp(0.0, 5.0);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Text(_status, style: theme.textTheme.bodyMedium),
        if (_cardWaitVisible) ...[
          const SizedBox(height: 12),
          Row(
            children: [
              Expanded(
                child: Text(
                  'Ожидание карты',
                  style: theme.textTheme.labelMedium,
                ),
              ),
              Text(
                '${elapsed.toStringAsFixed(1)} / 5 с',
                style: theme.textTheme.labelMedium?.copyWith(
                  color: theme.colorScheme.primary,
                  fontFeatures: const [FontFeature.tabularFigures()],
                ),
              ),
            ],
          ),
          const SizedBox(height: 6),
          LinearProgressIndicator(
            value: _cardWaitController.value,
            minHeight: 8,
            borderRadius: BorderRadius.circular(4),
          ),
        ],
      ],
    );
  }

  Widget _amountField({
    required TextEditingController controller,
    required String label,
  }) {
    return TextField(
      controller: controller,
      keyboardType: TextInputType.number,
      decoration: InputDecoration(
        labelText: label,
        border: const OutlineInputBorder(),
      ),
    );
  }

  Widget _tabScroll({required List<Widget> children}) {
    return ListView(
      padding: const EdgeInsets.all(16),
      children: children,
    );
  }

  Widget _primaryButton({
    required String label,
    required VoidCallback? onPressed,
    bool tonal = false,
  }) {
    final child = SizedBox(
      width: double.infinity,
      child: tonal
          ? FilledButton.tonal(onPressed: onPressed, child: Text(label))
          : FilledButton(onPressed: onPressed, child: Text(label)),
    );
    return child;
  }

  List<Widget> _cardSectionIfShown(int tabIndex) {
    if (_cardShownOnTab != tabIndex || _displayedPayload == null) {
      return const [];
    }
    return [
      const SizedBox(height: 24),
      CardPayloadView(payload: _displayedPayload),
    ];
  }

  Widget _tabRegister() {
    return _tabScroll(
      children: [
        const Text(
          'Запись на чип и автоматическая регистрация в БД. '
          'Номер карты = UID приложенной карты.',
        ),
        const SizedBox(height: 20),
        _amountField(
          controller: _initBalanceCtrl,
          label: 'Начальный баланс',
        ),
        const SizedBox(height: 20),
        _primaryButton(
          label: 'Зарегистрировать карту',
          tonal: true,
          onPressed: _busyOperation != null ? null : _initCardDefault,
        ),
        ..._cardSectionIfShown(_kTabRegister),
      ],
    );
  }

  Widget _tabInfo() {
    return _tabScroll(
      children: [
        const Text('Считать зашифрованные данные с приложенной карты.'),
        const SizedBox(height: 20),
        _primaryButton(
          label: 'Считать карту',
          tonal: true,
          onPressed: _busyOperation != null ? null : _readBalance,
        ),
        ..._cardSectionIfShown(_kTabInfo),
      ],
    );
  }

  Widget _tabDebit() {
    return _tabScroll(
      children: [
        const Text(
          'Списание: authorize на сервере, затем уменьшение balance на карте.',
        ),
        const SizedBox(height: 20),
        _amountField(
          controller: _debitAmountCtrl,
          label: 'Сумма списания',
        ),
        const SizedBox(height: 20),
        _primaryButton(
          label: 'Списать',
          onPressed: _busyOperation != null ? null : _debitTerminal,
        ),
        ..._cardSectionIfShown(_kTabDebit),
      ],
    );
  }

  Widget _tabCredit() {
    return _tabScroll(
      children: [
        const Text('Пополнение баланса на карте и событие на сервер.'),
        const SizedBox(height: 20),
        _amountField(
          controller: _creditAmountCtrl,
          label: 'Сумма пополнения',
        ),
        const SizedBox(height: 20),
        _primaryButton(
          label: 'Пополнить',
          onPressed: _busyOperation != null ? null : _creditMoney,
        ),
        ..._cardSectionIfShown(_kTabCredit),
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
              child: _terminalHeader(context),
            ),
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
              child: _statusBlock(context),
            ),
            TabBar(
              controller: _tabController,
              tabs: const [
                Tab(text: 'Регистрация'),
                Tab(text: 'Карта'),
                Tab(text: 'Списание'),
                Tab(text: 'Пополнение'),
              ],
            ),
            Expanded(
              child: TabBarView(
                controller: _tabController,
                children: [
                  _tabRegister(),
                  _tabInfo(),
                  _tabDebit(),
                  _tabCredit(),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
