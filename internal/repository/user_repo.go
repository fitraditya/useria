package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/fitraditya/useria/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *models.User) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, email, password_hash, first_name, last_name, oauth_provider, oauth_id, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		u.ID, u.Email, u.PasswordHash, u.FirstName, u.LastName, u.OAuthProvider, u.OAuthID, u.Status,
	)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return r.scanOne(r.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, first_name, last_name, avatar_url, oauth_provider, oauth_id, status, created_at, updated_at
		 FROM users WHERE email = ?`, email,
	))
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	return r.scanOne(r.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, first_name, last_name, avatar_url, oauth_provider, oauth_id, status, created_at, updated_at
		 FROM users WHERE id = ?`, id,
	))
}

func (r *UserRepository) UpdateProfile(ctx context.Context, id, firstName, lastName string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET first_name = ?, last_name = ?, updated_at = ? WHERE id = ?`,
		firstName, lastName, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("update user profile: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?`,
		passwordHash, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("update user password: %w", err)
	}
	return nil
}

// Count returns the total number of registered users.
func (r *UserRepository) Count(ctx context.Context) (int, error) {
	var n int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&n); err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return n, nil
}

func (r *UserRepository) scanOne(row *sql.Row) (*models.User, error) {
	var u models.User
	err := row.Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.AvatarURL,
		&u.OAuthProvider, &u.OAuthID, &u.Status, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return &u, nil
}
