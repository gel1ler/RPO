# RPO — NFC терминал (Flutter macOS)

## Сборка

Требования: **Xcode из App Store** или **Apple Command Line Tools**, Flutter с поддержкой desktop macOS.

```bash
brew install libnfc openssl@3   # см. xcconfig префикс Homebrew
flutter pub get
flutter run -d macos
```

Если сборка жалуется на пути `-lnfc` / OpenSSL — отредактируйте префиксы в `macos/Runner/Configs/NfcNative.xcconfig` под вашу платформу (ARM64 → `/opt/homebrew`, Intel → `/usr/local`).

Проект добавляет два **.c**-файла в Runner: `nfc_card.c` (чтение/запись, AES блоков), `mifare_compat.c` — обёртка `nfc_initiator_mifare_cmd` из утилит libnfc.

### Ошибка `unable to find utility "xcodebuild"`

Flutter для **macOS** всегда вызывает Xcode (Swift Package Manager в Runner и т.д.). Нужно:

1. Установить один из вариантов:
   - полный **Xcode** из App Store; первый запуск — открыть Xcode и принять лицензию;
   - либо только инструменты: в терминале выполнить **`xcode-select --install`** и дождаться окончания установки.

2. Указать активный developer directory (если стоит полный Xcode):

   ```bash
   sudo xcode-select -s /Applications/Xcode.app/Contents/Developer
   ```

3. Проверить:

   ```bash
   xcodebuild -version
   ```

4. Снова: `cd terminal && flutter doctor` (должен показать Xcode), затем `flutter run -d macos`.

Без `xcodebuild` проект **не соберётся** — это ограничение платформы, а не репозитория.

### Прочие сообщения Homebrew в терминале

- `No available formula with the name "#"` — обычно в команду `brew install` случайно попал символ **`#`** или обрезали строку; запускайте команды без комментариев в той же строке.
- ошибка **cask ezcast / `appcast`** — битый сторонний tap или устаревший cask; не связана с Flutter: обновите Homebrew (`brew update`) или удалите проблемный tap (`brew tap` / `brew untap ...`).

## Отладка без карты

`nfc_reader_present()` сообщает только о наличии устройства в списке libnfc; ошибки считывания видны после нажатия «Баланс с карты».

