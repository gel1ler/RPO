-- +goose Up
-- +goose StatementBegin
ALTER TABLE terminals ADD COLUMN api_secret_hash TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE terminals DROP COLUMN api_secret_hash;
-- +goose StatementEnd
