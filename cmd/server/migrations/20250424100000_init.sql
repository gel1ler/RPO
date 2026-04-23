-- +goose Up
-- +goose StatementBegin
CREATE TABLE terminals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    serial_number TEXT NOT NULL UNIQUE,
    address TEXT,
    name TEXT,
    extra TEXT,
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    label TEXT,
    key_value TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    card_number TEXT NOT NULL UNIQUE,
    balance INTEGER NOT NULL DEFAULT 0 CHECK (balance >= 0),
    is_blocked INTEGER NOT NULL DEFAULT 0 CHECK (is_blocked IN (0, 1)),
    owner_name TEXT,
    extra TEXT,
    key_id INTEGER NOT NULL REFERENCES keys(id) ON DELETE RESTRICT,
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE transactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    amount INTEGER NOT NULL CHECK (amount > 0),
    card_id INTEGER NOT NULL REFERENCES cards(id) ON DELETE RESTRICT,
    terminal_id INTEGER NOT NULL REFERENCES terminals(id) ON DELETE RESTRICT,
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    login TEXT NOT NULL UNIQUE,
    display_name TEXT,
    password_hash TEXT NOT NULL,
    is_admin INTEGER NOT NULL DEFAULT 0 CHECK (is_admin IN (0, 1)),
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_cards_key_id ON cards(key_id);
CREATE INDEX idx_transactions_card_id ON transactions(card_id);
CREATE INDEX idx_transactions_terminal_id ON transactions(terminal_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_transactions_terminal_id;
DROP INDEX IF EXISTS idx_transactions_card_id;
DROP INDEX IF EXISTS idx_cards_key_id;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS cards;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS terminals;
DROP TABLE IF EXISTS keys;
-- +goose StatementEnd
