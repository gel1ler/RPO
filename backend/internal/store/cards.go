package store

import (
	"context"
	"database/sql"
	"errors"
)

type Card struct {
	ID         int64
	CardNumber string
	Balance    int64
	IsBlocked  bool
	OwnerName  sql.NullString
	Extra      sql.NullString
	KeyID      int64
	CreatedAt  string
}

type CreateCardParams struct {
	CardNumber string
	Balance    int64
	IsBlocked  bool
	OwnerName  *string
	Extra      *string
	KeyID      int64
}

type UpdateCardParams struct {
	Balance   *int64
	IsBlocked *bool
	OwnerName *string
	Extra     *string
	KeyID     *int64
}

type Cards struct {
	DB *sql.DB
}

func (s Cards) GetByCardNumber(ctx context.Context, cardNumber string) (Card, error) {
	var c Card
	var isBlocked int64
	err := s.DB.QueryRowContext(ctx, `
SELECT id, card_number, balance, is_blocked, owner_name, extra, key_id, created_at
FROM cards
WHERE card_number = ?
`, cardNumber).Scan(&c.ID, &c.CardNumber, &c.Balance, &isBlocked, &c.OwnerName, &c.Extra, &c.KeyID, &c.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Card{}, ErrNotFound
	}
	if err != nil {
		return Card{}, err
	}
	c.IsBlocked = isBlocked == 1
	return c, nil
}

func (s Cards) List(ctx context.Context, limit int64, offset int64) ([]Card, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := s.DB.QueryContext(ctx, `
SELECT id, card_number, balance, is_blocked, owner_name, extra, key_id, created_at
FROM cards
ORDER BY id DESC
LIMIT ? OFFSET ?
`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Card
	for rows.Next() {
		var c Card
		var isBlocked int64
		if err := rows.Scan(&c.ID, &c.CardNumber, &c.Balance, &isBlocked, &c.OwnerName, &c.Extra, &c.KeyID, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.IsBlocked = isBlocked == 1
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s Cards) GetByID(ctx context.Context, id int64) (Card, error) {
	var c Card
	var isBlocked int64
	err := s.DB.QueryRowContext(ctx, `
SELECT id, card_number, balance, is_blocked, owner_name, extra, key_id, created_at
FROM cards
WHERE id = ?
`, id).Scan(&c.ID, &c.CardNumber, &c.Balance, &isBlocked, &c.OwnerName, &c.Extra, &c.KeyID, &c.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Card{}, ErrNotFound
	}
	if err != nil {
		return Card{}, err
	}
	c.IsBlocked = isBlocked == 1
	return c, nil
}

func (s Cards) Create(ctx context.Context, p CreateCardParams) (Card, error) {
	ownerName := sql.NullString{}
	if p.OwnerName != nil && *p.OwnerName != "" {
		ownerName = sql.NullString{String: *p.OwnerName, Valid: true}
	}
	extra := sql.NullString{}
	if p.Extra != nil && *p.Extra != "" {
		extra = sql.NullString{String: *p.Extra, Valid: true}
	}

	isBlocked := int64(0)
	if p.IsBlocked {
		isBlocked = 1
	}

	res, err := s.DB.ExecContext(ctx, `
INSERT INTO cards (card_number, balance, is_blocked, owner_name, extra, key_id)
VALUES (?, ?, ?, ?, ?, ?)
`, p.CardNumber, p.Balance, isBlocked, ownerName, extra, p.KeyID)
	if err != nil {
		return Card{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Card{}, err
	}
	return s.GetByID(ctx, id)
}

func (s Cards) Update(ctx context.Context, id int64, p UpdateCardParams) (Card, error) {
	c, err := s.GetByID(ctx, id)
	if err != nil {
		return Card{}, err
	}

	newBalance := c.Balance
	if p.Balance != nil {
		newBalance = *p.Balance
	}
	newIsBlocked := c.IsBlocked
	if p.IsBlocked != nil {
		newIsBlocked = *p.IsBlocked
	}
	newOwnerName := c.OwnerName
	if p.OwnerName != nil {
		if *p.OwnerName == "" {
			newOwnerName = sql.NullString{}
		} else {
			newOwnerName = sql.NullString{String: *p.OwnerName, Valid: true}
		}
	}
	newExtra := c.Extra
	if p.Extra != nil {
		if *p.Extra == "" {
			newExtra = sql.NullString{}
		} else {
			newExtra = sql.NullString{String: *p.Extra, Valid: true}
		}
	}
	newKeyID := c.KeyID
	if p.KeyID != nil {
		newKeyID = *p.KeyID
	}

	isBlocked := int64(0)
	if newIsBlocked {
		isBlocked = 1
	}

	_, err = s.DB.ExecContext(ctx, `
UPDATE cards
SET balance = ?, is_blocked = ?, owner_name = ?, extra = ?, key_id = ?
WHERE id = ?
`, newBalance, isBlocked, newOwnerName, newExtra, newKeyID, id)
	if err != nil {
		return Card{}, err
	}
	return s.GetByID(ctx, id)
}

func (s Cards) Delete(ctx context.Context, id int64) error {
	res, err := s.DB.ExecContext(ctx, `DELETE FROM cards WHERE id = ?`, id)
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
