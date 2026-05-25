---
name: ЛР4 Flutter NFC терминал
overview: Flutter macOS-терминал (FFI + libnfc), JSON на карте, backend authorize/keys + terminal_events/SSE, live-панель в админке.
todos:
  - id: user-prep
    content: Подготовка пользователя (выполнено)
    status: completed
  - id: flutter-scaffold
    content: terminal/ — конфиг, HTTP, FFI-каркас
    status: completed
  - id: native-nfc
    content: C/libnfc + decrypt/encrypt блоков + FFI
    status: completed
  - id: terminal-ui
    content: UI баланс / списание / пополнение balance и trips
    status: completed
  - id: backend-events
    content: terminal_events + POST event + SSE + лог authorize
    status: completed
  - id: frontend-live
    content: Live-панель операций терминалов
    status: completed
  - id: docs-acceptance
    content: README ЛР4
    status: pending
isProject: false
---

# ЛР4: Flutter-терминал NFC

## Зафиксировано

| Параметр | Значение |
|----------|----------|
| Терминал | `TERM-MAC-001` |
| Ключ | Key A, `FFFFFFFFFFFF` |
| Карта в БД | `1DFC7D05`, balance `1000`, key Key A |
| Суммы | `int` (как в backend) |
| Поездки | отдельное поле `trips` в JSON на карте |
| Платформа | macOS desktop |
| Backend | `https://localhost:8888/api/v1` |

## JSON на карте

```json
{
  "v": 1,
  "card_number": "1DFC7D05",
  "balance": 1000,
  "trips": 0,
  "key_id": 1
}
```

Сектор 1+, блоки по 16 байт, в C — `encrypt_block` / `decrypt_block` на каждый блок, ключ из `GET /terminal/keys`.

## Потоки

- **Баланс:** read card → показать `balance`, `trips`
- **Списание:** `POST /terminal/authorize` → при `approved` уменьшить `balance` на карте → write
- **Пополнение денег:** +`balance` на карте → write → `POST /terminal/event`
- **Пополнение поездок:** +`trips` на карте → write → event (без authorize)

Серверный `balance` в БД — только для `authorize`; перед демо держать равным карте.

## Реализация

1. `terminal/` — Flutter macOS, дефолты URL + `TERM-MAC-001`
2. `terminal/native/` — libnfc + crypto + FFI
3. Backend — `terminal_events`, `POST /terminal/event`, `GET /terminal/events/stream`, лог в `authorize`
4. Frontend — SSE на странице терминалов
5. README — запуск и приёмка

## Структура

```
RPO/terminal/     # Flutter + native/
RPO/backend/      # + events API
RPO/frontend/     # + live panel
```
