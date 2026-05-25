package store

import (
	"context"
	"database/sql"
	"errors"
)

type Terminal struct {
	ID           int64
	SerialNumber string
	Address      sql.NullString
	Name         sql.NullString
	Extra        sql.NullString
	CreatedAt    string
}

type CreateTerminalParams struct {
	SerialNumber string
	Address      *string
	Name         *string
	Extra        *string
}

type UpdateTerminalParams struct {
	Address *string
	Name    *string
	Extra   *string
}

type Terminals struct {
	DB *sql.DB
}

func (s Terminals) GetBySerialNumber(ctx context.Context, serialNumber string) (Terminal, error) {
	var t Terminal
	err := s.DB.QueryRowContext(ctx, `
SELECT id, serial_number, address, name, extra, created_at
FROM terminals
WHERE serial_number = ?
`, serialNumber).Scan(&t.ID, &t.SerialNumber, &t.Address, &t.Name, &t.Extra, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Terminal{}, ErrNotFound
	}
	if err != nil {
		return Terminal{}, err
	}
	return t, nil
}

func (s Terminals) List(ctx context.Context, limit int64, offset int64) ([]Terminal, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := s.DB.QueryContext(ctx, `
SELECT id, serial_number, address, name, extra, created_at
FROM terminals
ORDER BY id DESC
LIMIT ? OFFSET ?
`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Terminal
	for rows.Next() {
		var t Terminal
		if err := rows.Scan(&t.ID, &t.SerialNumber, &t.Address, &t.Name, &t.Extra, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s Terminals) GetByID(ctx context.Context, id int64) (Terminal, error) {
	var t Terminal
	err := s.DB.QueryRowContext(ctx, `
SELECT id, serial_number, address, name, extra, created_at
FROM terminals
WHERE id = ?
`, id).Scan(&t.ID, &t.SerialNumber, &t.Address, &t.Name, &t.Extra, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Terminal{}, ErrNotFound
	}
	if err != nil {
		return Terminal{}, err
	}
	return t, nil
}

func (s Terminals) Create(ctx context.Context, p CreateTerminalParams) (Terminal, error) {
	address := sql.NullString{}
	if p.Address != nil {
		address = sql.NullString{String: *p.Address, Valid: true}
	}
	name := sql.NullString{}
	if p.Name != nil {
		name = sql.NullString{String: *p.Name, Valid: true}
	}
	extra := sql.NullString{}
	if p.Extra != nil {
		extra = sql.NullString{String: *p.Extra, Valid: true}
	}

	res, err := s.DB.ExecContext(ctx, `
INSERT INTO terminals (serial_number, address, name, extra)
VALUES (?, ?, ?, ?)
`, p.SerialNumber, address, name, extra)
	if err != nil {
		return Terminal{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Terminal{}, err
	}
	return s.GetByID(ctx, id)
}

func (s Terminals) Update(ctx context.Context, id int64, p UpdateTerminalParams) (Terminal, error) {
	t, err := s.GetByID(ctx, id)
	if err != nil {
		return Terminal{}, err
	}

	newAddress := t.Address
	if p.Address != nil {
		if *p.Address == "" {
			newAddress = sql.NullString{}
		} else {
			newAddress = sql.NullString{String: *p.Address, Valid: true}
		}
	}
	newName := t.Name
	if p.Name != nil {
		if *p.Name == "" {
			newName = sql.NullString{}
		} else {
			newName = sql.NullString{String: *p.Name, Valid: true}
		}
	}
	newExtra := t.Extra
	if p.Extra != nil {
		if *p.Extra == "" {
			newExtra = sql.NullString{}
		} else {
			newExtra = sql.NullString{String: *p.Extra, Valid: true}
		}
	}

	_, err = s.DB.ExecContext(ctx, `
UPDATE terminals
SET address = ?, name = ?, extra = ?
WHERE id = ?
`, newAddress, newName, newExtra, id)
	if err != nil {
		return Terminal{}, err
	}
	return s.GetByID(ctx, id)
}

func (s Terminals) Delete(ctx context.Context, id int64) error {
	res, err := s.DB.ExecContext(ctx, `DELETE FROM terminals WHERE id = ?`, id)
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
