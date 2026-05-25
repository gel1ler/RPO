package store

import (
	"context"
	"database/sql"
	"errors"
)

type Key struct {
	ID        int64
	Label     sql.NullString
	KeyValue  string
	CreatedAt string
}

type CreateKeyParams struct {
	Label    *string
	KeyValue string
}

type UpdateKeyParams struct {
	Label    *string
	KeyValue *string
}

type Keys struct {
	DB *sql.DB
}

func (s Keys) List(ctx context.Context, limit int64, offset int64) ([]Key, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := s.DB.QueryContext(ctx, `
SELECT id, label, key_value, created_at
FROM keys
ORDER BY id DESC
LIMIT ? OFFSET ?
`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Key
	for rows.Next() {
		var k Key
		if err := rows.Scan(&k.ID, &k.Label, &k.KeyValue, &k.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s Keys) GetByID(ctx context.Context, id int64) (Key, error) {
	var k Key
	err := s.DB.QueryRowContext(ctx, `
SELECT id, label, key_value, created_at
FROM keys
WHERE id = ?
`, id).Scan(&k.ID, &k.Label, &k.KeyValue, &k.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Key{}, ErrNotFound
	}
	if err != nil {
		return Key{}, err
	}
	return k, nil
}

func (s Keys) Create(ctx context.Context, p CreateKeyParams) (Key, error) {
	label := sql.NullString{}
	if p.Label != nil && *p.Label != "" {
		label = sql.NullString{String: *p.Label, Valid: true}
	}

	res, err := s.DB.ExecContext(ctx, `
INSERT INTO keys (label, key_value)
VALUES (?, ?)
`, label, p.KeyValue)
	if err != nil {
		return Key{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Key{}, err
	}
	return s.GetByID(ctx, id)
}

func (s Keys) Update(ctx context.Context, id int64, p UpdateKeyParams) (Key, error) {
	k, err := s.GetByID(ctx, id)
	if err != nil {
		return Key{}, err
	}

	newLabel := k.Label
	if p.Label != nil {
		if *p.Label == "" {
			newLabel = sql.NullString{}
		} else {
			newLabel = sql.NullString{String: *p.Label, Valid: true}
		}
	}

	newKeyValue := k.KeyValue
	if p.KeyValue != nil {
		newKeyValue = *p.KeyValue
	}

	_, err = s.DB.ExecContext(ctx, `
UPDATE keys
SET label = ?, key_value = ?
WHERE id = ?
`, newLabel, newKeyValue, id)
	if err != nil {
		return Key{}, err
	}
	return s.GetByID(ctx, id)
}

func (s Keys) Delete(ctx context.Context, id int64) error {
	res, err := s.DB.ExecContext(ctx, `DELETE FROM keys WHERE id = ?`, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}
