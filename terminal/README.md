# RPO — NFC терминал (Flutter macOS)

## Сборка

Требования: **Xcode из App Store** или **Apple Command Line Tools**, Flutter с поддержкой desktop macOS.

```bash
brew install libnfc openssl@3
```

Проверьте, что каталоги существуют (иначе не найдётся `nfc/nfc.h`):

```bash
ls "$(brew --prefix libnfc)/include/nfc/nfc.h"
```

По умолчанию в `NfcNative.xcconfig` заданы пути **Apple Silicon** (`/opt/homebrew/opt/...`). На **Intel Homebrew** добавьте в `HEADER_SEARCH_PATHS` и `LIBRARY_SEARCH_PATHS` каталоги `/usr/local/opt/libnfc/...` и `/usr/local/opt/openssl@3/...` (см. комментарий в файле).

```bash
flutter pub get
flutter run -d macos
```

Проект добавляет два **.c**-файла в Runner: `nfc_card.c` (чтение/запись, AES блоков), `mifare_compat.c` — обёртка `nfc_initiator_mifare_cmd` из утилит libnfc.

### Ошибка `unable to find utility "xcodebuild"`

Flutter для **macOS** всегда вызывает Xcode (Swift Package Manager в Runner и т.д.). Нужно:

1. Установить **полный Xcode** из App Store; первый запуск — открыть Xcode и принять лицензию (**только** Command Line Tools без `Xcode.app` для `flutter run -d macos` недостаточно).

2. Указать активный developer directory и довести Xcode до рабочего состояния:

   ```bash
   sudo xcode-select -s /Applications/Xcode.app/Contents/Developer
   sudo xcodebuild -license
   sudo xcodebuild -runFirstLaunch
   ```

3. Проверить:

   ```bash
   xcodebuild -version
   ```

4. Снова: `cd terminal && flutter doctor`, затем `flutter run -d macos`.

Без `xcodebuild` проект **не соберётся** — это ограничение платформы, а не репозитория.

### Прочие сообщения Homebrew в терминале

- `No available formula with the name "#"` — обычно в команду `brew install` случайно попал символ **`#`** или обрезали строку; запускайте команды без комментариев в той же строке.
- ошибка **cask ezcast / `appcast`** — битый сторонний tap или устаревший cask; не связана с Flutter: обновите Homebrew (`brew update`) или удалите проблемный tap (`brew tap` / `brew untap ...`).

## Авторизация на API

При запуске приложение логинится: `POST /api/v1/terminal/auth/login` с `serial_number` и `api_secret` из `lib/terminal_config.dart` (по умолчанию совпадает с `APP_TERMINAL_API_SECRET` в `docker-compose.yml`). Полученный JWT передаётся в заголовке `Authorization` на `/terminal/authorize`, `/terminal/keys`, `/terminal/register-card`, `/terminal/event`.

## Отладка без карты

`nfc_reader_present()` сообщает только о наличии устройства в списке libnfc; ошибки считывания видны после нажатия «Баланс с карты».

