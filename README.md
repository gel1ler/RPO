# RPO (лабораторные)

Стек: **backend** Go + SQLite в `backend/`, фронт **React (Vite + MUI)** в `frontend/`, один образ Docker (Go + сборка SPA + nginx + TLS).

## Запуск через Docker

```bash
docker compose up --build
```

- Интерфейс и API: `https://localhost:8888`
- По умолчанию в compose задан админ: `admin` / `admin` (самоподписанный сертификат — браузеру нужно принять исключение, для curl — `-k`)

**Карты:** у администратора поле «Владелец» — выпадающий список пользователей (в БД по-прежнему сохраняется строка `owner_name`: отображаемое имя или логин). У обычного пользователя — текстовое поле (нет доступа к списку `/users`).

**Транзакции и баланс:** в задании разделены (1) CRUD таблицы транзакций — учёт записей в БД, (2) `POST /terminal/authorize` — только **решение** об одобрении платежа по карте и сумме, без обязательного списания. Текущий бэкенд при создании транзакции через API **не изменяет** `balance` карты; это согласуется с тем, что п. 5.1 описывает именно авторизацию для терминала, а не модель биллинга. Если преподаватель требует списание — это отдельное правило (например, уменьшать баланс при успешном `authorize` или при `POST /transactions`).

## Локальная разработка фронта

1. Запустить бэк (compose или из `backend/`: `go run ./cmd/server`).
2. `cd frontend && npm install && npm run dev` — прокси к API настроен в `vite.config.ts` (`/api` → `https://localhost:8888`).

Сборка бэкенда без Docker:

```bash
cd backend && go build -o server ./cmd/server
```

## ЛР4 — Flutter-терминал NFC (macOS)

Приложение в каталоге [`terminal`](terminal/README.md):

- **Flutter macOS** + нативный **C**: libnfc, MIFARE Classic (данные в блоках 4–6 и 8–10), **AES-128-ECB** на каждый 16-байтный блок, JSON на карте: `{ v, card_number, balance, trips, key_id }`.
- Расширение API ЛР2: `POST /api/v1/terminal/event`, `GET /api/v1/terminal/events`; каждый вызов `POST /terminal/authorize` сохраняется в `terminal_events`.
- Страница **Терминалы** в SPA показывает **журнал** с опросом раз в ~2 с.

Подготовка (Homebrew, Apple Silicon; на Intel см. префикс в [`terminal/macos/Runner/Configs/NfcNative.xcconfig`](terminal/macos/Runner/Configs/NfcNative.xcconfig)):

```bash
brew install libnfc openssl@3
```

На машине с установленным **Xcode** или **Command Line Tools** (`xcodebuild` в PATH):

```bash
cd terminal && flutter pub get && flutter run -d macos
```

Если видите `unable to find utility "xcodebuild"` — см. раздел устранения в [`terminal/README.md`](terminal/README.md).

По умолчанию приложение ходит в `https://127.0.0.1:8888/api/v1` и принимает self-signed TLS. После операций админ может увидеть события в журнале.

**Приёмка:** `docker compose up`, «Записать карту» в приложении; баланс карты на чипе и в админке для демонстрации `authorize` нужно держать **согласованными** (сервер только решает, не изменяет баланс автоматически).
