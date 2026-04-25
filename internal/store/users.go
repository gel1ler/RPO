package store

import (
	"context"
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("not found")

type User struct {
	ID           int64
	Login        string
	DisplayName  sql.NullString
	PasswordHash string
	IsAdmin      bool
	CreatedAt    string
}

type CreateUserParams struct {
	Login        string
	DisplayName  *string
	PasswordHash string
	IsAdmin      bool
}

type UpdateUserParams struct {
	DisplayName  *string
	PasswordHash *string
	IsAdmin      *bool
}

type Users struct {
	DB *sql.DB
}

func (s Users) GetByID(ctx context.Context, id int64) (User, error) {
	var u User
	var isAdmin int64
	err := s.DB.QueryRowContext(ctx, `
SELECT id, login, display_name, password_hash, is_admin, created_at
FROM users
WHERE id = ?
`, id).Scan(&u.ID, &u.Login, &u.DisplayName, &u.PasswordHash, &isAdmin, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	u.IsAdmin = isAdmin == 1
	return u, nil
}

func (s Users) GetByLogin(ctx context.Context, login string) (User, error) {
	var u User
	var isAdmin int64
	err := s.DB.QueryRowContext(ctx, `
SELECT id, login, display_name, password_hash, is_admin, created_at
FROM users
WHERE login = ?
`, login).Scan(&u.ID, &u.Login, &u.DisplayName, &u.PasswordHash, &isAdmin, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	u.IsAdmin = isAdmin == 1
	return u, nil
}

func (s Users) List(ctx context.Context, limit int64, offset int64) ([]User, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := s.DB.QueryContext(ctx, `
SELECT id, login, display_name, password_hash, is_admin, created_at
FROM users
ORDER BY id DESC
LIMIT ? OFFSET ?
`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []User
	for rows.Next() {
		var u User
		var isAdmin int64
		if err := rows.Scan(&u.ID, &u.Login, &u.DisplayName, &u.PasswordHash, &isAdmin, &u.CreatedAt); err != nil {
			return nil, err
		}
		u.IsAdmin = isAdmin == 1
		out = append(out, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s Users) Create(ctx context.Context, p CreateUserParams) (User, error) {
	displayName := sql.NullString{}
	if p.DisplayName != nil {
		displayName = sql.NullString{String: *p.DisplayName, Valid: true}
	}
	isAdmin := int64(0)
	if p.IsAdmin {
		isAdmin = 1
	}

	res, err := s.DB.ExecContext(ctx, `
INSERT INTO users (login, display_name, password_hash, is_admin)
VALUES (?, ?, ?, ?)
`, p.Login, displayName, p.PasswordHash, isAdmin)
	if err != nil {
		return User{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return User{}, err
	}
	return s.GetByID(ctx, id)
}

func (s Users) Update(ctx context.Context, id int64, p UpdateUserParams) (User, error) {
	u, err := s.GetByID(ctx, id)
	if err != nil {
		return User{}, err
	}

	newDisplayName := u.DisplayName
	if p.DisplayName != nil {
		if *p.DisplayName == "" {
			newDisplayName = sql.NullString{}
		} else {
			newDisplayName = sql.NullString{String: *p.DisplayName, Valid: true}
		}
	}
	newPasswordHash := u.PasswordHash
	if p.PasswordHash != nil {
		newPasswordHash = *p.PasswordHash
	}
	newIsAdmin := u.IsAdmin
	if p.IsAdmin != nil {
		newIsAdmin = *p.IsAdmin
	}

	isAdmin := int64(0)
	if newIsAdmin {
		isAdmin = 1
	}

	_, err = s.DB.ExecContext(ctx, `
UPDATE users
SET display_name = ?, password_hash = ?, is_admin = ?
WHERE id = ?
`, newDisplayName, newPasswordHash, isAdmin, id)
	if err != nil {
		return User{}, err
	}
	return s.GetByID(ctx, id)
}

func (s Users) Delete(ctx context.Context, id int64) error {
	res, err := s.DB.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
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

func (s Users) EnsureAdmin(ctx context.Context, login string, passwordHash string) (User, error) {
	u, err := s.GetByLogin(ctx, login)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return User{}, err
	}
	if errors.Is(err, ErrNotFound) {
		return s.Create(ctx, CreateUserParams{
			Login:        login,
			DisplayName:  nil,
			PasswordHash: passwordHash,
			IsAdmin:      true,
		})
	}

	isAdmin := true
	return s.Update(ctx, u.ID, UpdateUserParams{
		PasswordHash: &passwordHash,
		IsAdmin:      &isAdmin,
	})
}
