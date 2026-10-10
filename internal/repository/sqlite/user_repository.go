package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"famloop/internal/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	if u.Avatar == "" {
		u.Avatar = domain.GetRandomDefaultAvatar()
	}

	query := `
		INSERT INTO users (id, username, email, phone_number, avatar, password_hash, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now().UTC()
	u.CreatedAt = now
	u.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query, u.ID, u.Username, u.Email, u.PhoneNumber, u.Avatar, u.PasswordHash, u.CreatedAt, u.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `SELECT id, username, email, phone_number, avatar, password_hash, created_at, updated_at FROM users WHERE id = ?`
	return r.scanUser(r.db.QueryRowContext(ctx, query, id))
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, username, email, phone_number, avatar, password_hash, created_at, updated_at FROM users WHERE email = ?`
	return r.scanUser(r.db.QueryRowContext(ctx, query, email))
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `SELECT id, username, email, phone_number, avatar, password_hash, created_at, updated_at FROM users WHERE username = ?`
	return r.scanUser(r.db.QueryRowContext(ctx, query, username))
}

func (r *UserRepository) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	query := `SELECT id, username, email, phone_number, avatar, password_hash, created_at, updated_at FROM users WHERE phone_number = ?`
	return r.scanUser(r.db.QueryRowContext(ctx, query, phone))
}

func (r *UserRepository) UpdateAvatar(ctx context.Context, id, avatar string) error {
	query := `UPDATE users SET avatar = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, avatar, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to update user avatar: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, u *domain.User) error {
	query := `UPDATE users SET username = ?, email = ?, phone_number = ?, updated_at = ? WHERE id = ?`
	u.UpdatedAt = time.Now().UTC()
	_, err := r.db.ExecContext(ctx, query, u.Username, u.Email, u.PhoneNumber, u.UpdatedAt, u.ID)
	if err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}
	return nil
}

func (r *UserRepository) scanUser(row *sql.Row) (*domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PhoneNumber, &u.Avatar, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan user failed: %w", err)
	}
	return &u, nil
}
