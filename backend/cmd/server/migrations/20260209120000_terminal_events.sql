-- +goose Up
-- +goose StatementBegin
CREATE TABLE terminal_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    terminal_serial TEXT NOT NULL,
    card_number TEXT NOT NULL,
    operation TEXT NOT NULL,
    amount INTEGER NOT NULL DEFAULT 0 CHECK (amount >= 0),
    trips_delta INTEGER NOT NULL DEFAULT 0,
    approved INTEGER CHECK (approved IS NULL OR approved IN (0, 1)),
    reason TEXT,
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_terminal_events_id ON terminal_events(id);
CREATE INDEX idx_terminal_events_created_at ON terminal_events(created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_terminal_events_created_at;
DROP INDEX IF EXISTS idx_terminal_events_id;
DROP TABLE IF EXISTS terminal_events;
-- +goose StatementEnd
