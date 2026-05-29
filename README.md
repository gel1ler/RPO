# RPO (лабораторные)

Стек: **backend** Go + SQLite в `backend/`, фронт **React (Vite + MUI)** в `frontend/`, один образ Docker (Go + сборка SPA + nginx + TLS).

## Запуск через Docker

```bash
docker compose up --build
```

- Интерфейс и API: `https://localhost:8888`
- Swagger: `https://localhost:8888/api/v1/swagger/index.html`
- По умолчанию в compose задан админ: `admin` / `admin` (самоподписанный сертификат — браузеру нужно принять исключение, для curl — `-k`)

**Авторизация в Swagger:** `POST /auth/login` → скопировать поле `token` → **Authorize** → в `Authorization` вставить `Bearer <token>` (со словом Bearer и пробелом) или только сам JWT. Нажать **Authorize**, затем **Close**. Если вставить `Bearer Bearer …`, будет 401.

**Карты:** у администратора поле «Владелец» — выпадающий список пользователей (в БД по-прежнему сохраняется строка `owner_name`: отображаемое имя или логин). У обычного пользователя — текстовое поле (нет доступа к списку `/users`).

**Транзакции и баланс:** `POST /terminal/authorize` — только **решение** об одобрении платежа. Списание и пополнение с терминала (`debit_card` / `credit_balance` через `POST /terminal/event`) создают запись в `transactions` и атомарно меняют `balance` карты. CRUD `POST /api/v1/transactions` (админка) по-прежнему только добавляет запись без изменения баланса.

## Локальная разработка фронта

1. Запустить бэк (compose или из `backend/`: `go run ./cmd/server`).
2. `cd frontend && npm install && npm run dev` — прокси к API настроен в `vite.config.ts` (`/api` → `https://localhost:8888`).

Сборка бэкенда без Docker:

```bash
cd backend && go build -o server ./cmd/server
```

## ЛР4 — Flutter-терминал NFC (macOS)

Приложение в каталоге [`terminal`](terminal/README.md):

- **Flutter macOS** + нативный **C**: libnfc, MIFARE Classic (данные в блоках 4–6 и 8–10), **AES-128-ECB** на каждый 16-байтный блок, JSON на карте: `{ v, card_number, balance, key_id }`.
- Расширение API ЛР2: `POST /api/v1/terminal/register-card`, `POST /api/v1/terminal/event`, `GET /api/v1/terminal/events`; каждый вызов `POST /terminal/authorize` сохраняется в `terminal_events`.
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

По умолчанию приложение ходит в `https://127.0.0.1:8888/api/v1`, при старте выполняет `POST /terminal/auth/login` (serial `TERM-MAC-001`, secret `terminal-dev-secret` — как в `docker-compose.yml` и `terminal/lib/terminal_config.dart`) и дальше шлёт `Authorization: Bearer …` на все `/terminal/*`, кроме login.

После операций админ может увидеть события в журнале. Новый терминал в админке создаётся с полем `api_secret` в ответе (один раз) — его нужно прописать в конфиге Flutter-терминала.

**Приёмка:** `docker compose up`, «Записать карту» в приложении; баланс карты на чипе и в админке для демонстрации `authorize` нужно держать **согласованными** (сервер только решает, не изменяет баланс автоматически).
