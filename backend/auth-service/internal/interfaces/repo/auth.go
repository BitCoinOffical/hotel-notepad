package repo

import (
	"context"
	"fmt"

	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/domain/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepo struct {
	pool *pgxpool.Pool
}

func NewAuthRepo(pool *pgxpool.Pool) *AuthRepo {
	return &AuthRepo{pool: pool}
}

func (r *AuthRepo) SaveUser(ctx context.Context, user *models.User) (string, error) {
	sql := `INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id`
	var userID string
	err := r.pool.QueryRow(ctx, sql, user.UserName, user.PasswordHash).Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("r.pool.QueryRow: %w", err)
	}
	return userID, nil
}

func (r *AuthRepo) GetUserByUserName(ctx context.Context, username string) (*models.User, error) {
	sql := `SELECT id, username, password_hash, role FROM users WHERE username = $1`
	var user models.User
	if err := r.pool.QueryRow(ctx, sql, username).Scan(&user); err != nil {
		return nil, fmt.Errorf("r.pool.QueryRow: %w", err)
	}
	return &user, nil
}

func (r *AuthRepo) CheckUserExists(ctx context.Context, username string) (bool, error) {
	sql := `SELECT EXISTS (SELECT 1 FROM users WHERE username = $1)`
	var exists bool
	if err := r.pool.QueryRow(ctx, sql, username).Scan(&exists); err != nil {
		return false, fmt.Errorf("r.pool.QueryRow: %w", err)
	}
	return exists, nil
}
