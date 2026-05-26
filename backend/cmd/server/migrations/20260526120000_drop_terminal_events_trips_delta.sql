-- +goose Up
-- +goose StatementBegin
ALTER TABLE terminal_events DROP COLUMN trips_delta;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE terminal_events ADD COLUMN trips_delta INTEGER NOT NULL DEFAULT 0;
-- +goose StatementEnd
