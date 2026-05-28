import 'dart:convert';
import 'dart:ffi' as ffi;
import 'dart:io';
import 'dart:isolate';

import 'package:ffi/ffi.dart';

final ffi.DynamicLibrary _dylib = () {
  if (Platform.isMacOS) return ffi.DynamicLibrary.executable();
  throw UnsupportedError('Терминал поддерживается только на macOS');
}();

ffi.Pointer<Utf8> _allocUtf8(String s) => s.toNativeUtf8();

String _peekUtf8Zeros(ffi.Pointer<ffi.Uint8> p, int maxBytes) {
  final raw = <int>[];
  for (var i = 0; i < maxBytes; i++) {
    final v = (p + i).value;
    if (v == 0) break;
    raw.add(v);
  }
  return utf8.decode(raw);
}

class NfcBridge {
  NfcBridge._();

  static final _nfcReaderPresent = _dylib
      .lookup<ffi.NativeFunction<ffi.Int32 Function()>>('nfc_reader_present')
      .asFunction<int Function()>();

  static final _nfcReadCard = _dylib
      .lookup<
          ffi.NativeFunction<
              ffi.Int32 Function(
                ffi.Pointer<Utf8>,
                ffi.Pointer<ffi.Uint8>,
                ffi.IntPtr,
                ffi.Pointer<ffi.Uint8>,
                ffi.IntPtr,
              )>>('nfc_read_card')
      .asFunction<
          int Function(
            ffi.Pointer<Utf8>,
            ffi.Pointer<ffi.Uint8>,
            int,
            ffi.Pointer<ffi.Uint8>,
            int,
          )>();

  static final _nfcReadUid = _dylib
      .lookup<
          ffi.NativeFunction<
              ffi.Int32 Function(
                ffi.Pointer<ffi.Uint8>,
                ffi.IntPtr,
              )>>('nfc_read_uid')
      .asFunction<
          int Function(
            ffi.Pointer<ffi.Uint8>,
            int,
          )>();

  static final _nfcWriteCard = _dylib
      .lookup<
          ffi.NativeFunction<
              ffi.Int32 Function(
                ffi.Pointer<Utf8>,
                ffi.Pointer<Utf8>,
              )>>('nfc_write_card')
      .asFunction<
          int Function(
            ffi.Pointer<Utf8>,
            ffi.Pointer<Utf8>,
          )>();

  static final _nfcLastError = _dylib
      .lookup<ffi.NativeFunction<ffi.Pointer<Utf8> Function()>>('nfc_last_error')
      .asFunction<ffi.Pointer<Utf8> Function()>();

  static bool readerPresent() => _nfcReaderPresent() != 0;

  /// Сообщение из [nfc_last_error] после ошибки FFI.
  static String lastLibError() {
    final ptr = _nfcLastError();
    if (ptr.address == 0) return '';
    return ptr.toDartString();
  }

  /// [mifareKeyHex]: 12 hex-символов, напр. FFFFFFFFFFFF.
  static ({String uid, String json}) readCardPayload(String mifareKeyHex) {
    final keyPtr = _allocUtf8(mifareKeyHex);
    final uidBuf = calloc<ffi.Uint8>(64);
    const jsonCapacity = 1024;
    final jsonBuf = calloc<ffi.Uint8>(jsonCapacity);
    try {
      final rc = _nfcReadCard(keyPtr, uidBuf, 64, jsonBuf, jsonCapacity);
      if (rc != 0) {
        throw FormatException(lastLibError());
      }
      return (
        uid: _peekUtf8Zeros(uidBuf, 64),
        json: _peekUtf8Zeros(jsonBuf, jsonCapacity),
      );
    } finally {
      calloc.free(keyPtr);
      calloc.free(uidBuf);
      calloc.free(jsonBuf);
    }
  }

  static String readUidOnly() {
    final uidBuf = calloc<ffi.Uint8>(64);
    try {
      final rc = _nfcReadUid(uidBuf, 64);
      if (rc != 0) {
        throw FormatException(lastLibError());
      }
      return _peekUtf8Zeros(uidBuf, 64);
    } finally {
      calloc.free(uidBuf);
    }
  }

  static void writeCardPayload(String mifareKeyHex, String json) {
    final keyPtr = _allocUtf8(mifareKeyHex);
    final jsonPtr = _allocUtf8(json);
    try {
      final rc = _nfcWriteCard(keyPtr, jsonPtr);
      if (rc != 0) {
        throw FormatException(lastLibError());
      }
    } finally {
      calloc.free(keyPtr);
      calloc.free(jsonPtr);
    }
  }

  /// NFC в фоне — UI остаётся отзывчивым, анимация ожидания карты идёт на main isolate.
  static Future<({String uid, String json})> readCardPayloadAsync(
    String mifareKeyHex,
  ) {
    return Isolate.run(() => readCardPayload(mifareKeyHex));
  }

  static Future<void> writeCardPayloadAsync(
    String mifareKeyHex,
    String json,
  ) {
    return Isolate.run(() => writeCardPayload(mifareKeyHex, json));
  }

  static Future<String> readUidOnlyAsync() {
    return Isolate.run(readUidOnly);
  }
}
