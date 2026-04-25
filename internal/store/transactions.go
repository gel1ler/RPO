package store

import (
	"context"
	"database/sql"
	"errors"
)

type Transaction struct {
	ID         int64
	Amount     int64
	CardID     int64
	TerminalID int64
	CreatedAt  string
}

type CreateTransactionParams struct {
	Amount     int64
	CardID     int64
	TerminalID int64
}

type Transactions struct {
	DB *sql.DB
}

func (s Transactions) List(ctx context.Context, limit int64, offset int64) ([]Transaction, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := s.DB.QueryContext(ctx, `
SELECT id, amount, card_id, terminal_id, created_at
FROM transactions
ORDER BY id DESC
LIMIT ? OFFSET ?
`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Transaction
	for rows.Next() {
		var t Transaction
		if err := rows.Scan(&t.ID, &t.Amount, &t.CardID, &t.TerminalID, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s Transactions) GetByID(ctx context.Context, id int64) (Transaction, error) {
	var t Transaction
	err := s.DB.QueryRowContext(ctx, `
SELECT id, amount, card_id, terminal_id, created_at
FROM transactions
WHERE id = ?
`, id).Scan(&t.ID, &t.Amount, &t.CardID, &t.TerminalID, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Transaction{}, ErrNotFound
	}
	if err != nil {
		return Transaction{}, err
	}
	return t, nil
}

func (s Transactions) Create(ctx context.Context, p CreateTransactionParams) (Transaction, error) {
	res, err := s.DB.ExecContext(ctx, `
INSERT INTO transactions (amount, card_id, terminal_id)
VALUES (?, ?, ?)
`, p.Amount, p.CardID, p.TerminalID)
	if err != nil {
		return Transaction{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Transaction{}, err
	}
	return s.GetByID(ctx, id)
}

func (s Transactions) Delete(ctx context.Context, id int64) error {
	res, err := s.DB.ExecContext(ctx, `DELETE FROM transactions WHERE id = ?`, id)
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
