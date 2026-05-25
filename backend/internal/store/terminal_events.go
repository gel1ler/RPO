package store

import (
	"context"
	"database/sql"
	"errors"
)

type TerminalEvent struct {
	ID               int64
	TerminalSerial   string
	CardNumber       string
	Operation        string
	Amount           int64
	TripsDelta       int64
	Approved         sql.NullInt64 // NULL = не применимо; 0/1
	Reason           sql.NullString
	CreatedAt        string
}

type CreateTerminalEventParams struct {
	TerminalSerial string
	CardNumber     string
	Operation      string
	Amount         int64
	TripsDelta     int64
	Approved       *bool
	Reason         *string
}

type TerminalEvents struct {
	DB *sql.DB
}

func (s TerminalEvents) Create(ctx context.Context, p CreateTerminalEventParams) (TerminalEvent, error) {
	var approvedIface interface{}
	if p.Approved != nil {
		v := int64(0)
		if *p.Approved {
			v = 1
		}
		approvedIface = v
	}

	var reasonIface interface{}
	if p.Reason != nil {
		reasonIface = *p.Reason
	}

	res, err := s.DB.ExecContext(ctx, `
INSERT INTO terminal_events (terminal_serial, card_number, operation, amount, trips_delta, approved, reason)
VALUES (?,?,?,?,?,?,?)`,
		p.TerminalSerial, p.CardNumber, p.Operation, p.Amount, p.TripsDelta, approvedIface, reasonIface,
	)
	if err != nil {
		return TerminalEvent{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return TerminalEvent{}, err
	}
	return s.GetByID(ctx, id)
}

func (s TerminalEvents) GetByID(ctx context.Context, id int64) (TerminalEvent, error) {
	row := s.DB.QueryRowContext(ctx, `
SELECT id, terminal_serial, card_number, operation, amount, trips_delta, approved, reason, created_at
FROM terminal_events WHERE id = ?`, id)
	return scanTerminalEvent(row)
}

func (s TerminalEvents) ListSince(ctx context.Context, sinceID int64, limit int64) ([]TerminalEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	rows, err := s.DB.QueryContext(ctx, `
SELECT id, terminal_serial, card_number, operation, amount, trips_delta, approved, reason, created_at
FROM terminal_events WHERE id > ? ORDER BY id ASC LIMIT ?`, sinceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []TerminalEvent{}
	for rows.Next() {
		e, err := scanTerminalEventRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func scanTerminalEvent(row *sql.Row) (TerminalEvent, error) {
	var e TerminalEvent
	err := row.Scan(&e.ID, &e.TerminalSerial, &e.CardNumber, &e.Operation, &e.Amount, &e.TripsDelta, &e.Approved, &e.Reason, &e.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return TerminalEvent{}, ErrNotFound
		}
		return TerminalEvent{}, err
	}
	return e, nil
}

func scanTerminalEventRow(rows *sql.Rows) (TerminalEvent, error) {
	var e TerminalEvent
	err := rows.Scan(&e.ID, &e.TerminalSerial, &e.CardNumber, &e.Operation, &e.Amount, &e.TripsDelta, &e.Approved, &e.Reason, &e.CreatedAt)
	return e, err
}
