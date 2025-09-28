package db

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// User represents an application user.
// PasswordHash stores argon2id hash from security.HashPassword.
type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

var ensureUsersDDL = `
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

// EnsureUsersTable ensures the users table exists.
func EnsureUsersTable(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return errors.New("no database pool")
	}
	_, err := pool.Exec(ctx, ensureUsersDDL)
	return err
}

// CreateUser inserts a new user.
func CreateUser(ctx context.Context, pool *pgxpool.Pool, email, passwordHash string) (*User, error) {
	if pool == nil {
		return nil, errors.New("no database pool")
	}
	email = strings.TrimSpace(strings.ToLower(email))
	id := uuid.New()
	var u User
	err := pool.QueryRow(ctx,
		`INSERT INTO users (id, email, password_hash) VALUES ($1,$2,$3)
		RETURNING id, email, password_hash, created_at, updated_at`,
		id, email, passwordHash,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetUserByEmail returns the user by email or pgx.ErrNoRows.
func GetUserByEmail(ctx context.Context, pool *pgxpool.Pool, email string) (*User, error) {
	if pool == nil {
		return nil, errors.New("no database pool")
	}
	email = strings.TrimSpace(strings.ToLower(email))
	var u User
	err := pool.QueryRow(ctx,
		`SELECT id, email, password_hash, created_at, updated_at FROM users WHERE email=$1`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
